/**------------------------------------------------------------**
 * @filename rabbitmq/config.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-06 13:12
 * @desc     go.jd100.com - rabbitmq - config
 **------------------------------------------------------------**/
package rabbitmq

type Config struct {
	Dsn string
	Vhost string

	Locale string
	Exchange *ExchangeConfig
}


type ExchangeConfig struct {
	Name       string
	Type       string
	RoutingKey string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Queue      *QueueConfig
}

type QueueConfig struct {
	Name  string
}