/**------------------------------------------------------------**
 * @filename rabbitmq/rabbit.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-06 13:11
 * @desc     go.jd100.com - rabbitmq - rabbit
 **------------------------------------------------------------**/
package rabbitmq

import (
	"fmt"

	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/queue/amqp"
)


func Publish(c *Config, body string) error {
	conn, err := amqp.Dial(c.Dsn)
	if err != nil {
		log.Errorf("Dial: %v", err)
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Errorf("Channel: %v", err)
		return err
	}
	if err := channel.ExchangeDeclare(
		c.Exchange.Name,     // name
		c.Exchange.Type, // type
		c.Exchange.Durable,         // durable
		c.Exchange.AutoDelete,        // auto-deleted
		c.Exchange.Internal,        // internal
		c.Exchange.NoWait,        // noWait
		nil,          // arguments
	); err != nil {
		log.Errorf("Exchange Declare: %v", err)
		return err
	}
	if err = channel.Publish(
		c.Exchange.Name,   // publish to an exchange
		c.Exchange.Queue.Name,        // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Errorf("Exchange Publish: %s", err)
		return err
	}
	return nil
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(cfg *Config, key, ctag string) (c *Consumer, err error) {
	c = &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	c.conn, err = amqp.Dial(cfg.Dsn)
	if err != nil {
		log.Errorf("Dial: %v", err)
		return nil, err
	}
	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}
	if err = c.channel.ExchangeDeclare(
		cfg.Exchange.Name,     // name of the exchange
		cfg.Exchange.Type, // type
		cfg.Exchange.Durable,         // durable
		cfg.Exchange.AutoDelete,        // auto-deleted
		cfg.Exchange.Internal,        // internal
		cfg.Exchange.NoWait,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}
	queue, err := c.channel.QueueDeclare(
		cfg.Exchange.Queue.Name, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		cfg.Exchange.Name,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Info("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		log.Info(string(d.Body))
		d.Ack(false)
	}
	done <- nil
}
