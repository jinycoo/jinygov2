/**------------------------------------------------------------**
 * @filename redis/redis.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-19 11:56
 * @desc     go.jd100.com - redis - redis
 **------------------------------------------------------------**/
package redis

import (
	"context"
	"fmt"
	"time"

	"go.jd100.com/medusa/cache/redis/internal"
	"go.jd100.com/medusa/cache/redis/internal/pool"
	"go.jd100.com/medusa/cache/redis/internal/proto"
	"go.jd100.com/medusa/net/netutil/breaker"
)

// Nil reply Redis returns when key does not exist.
const Nil = proto.Nil

type baseClient struct {
	cfg      *Config
	connPool pool.Pooler
	limiter  breaker.Breaker

	onClose func() error // hook called when client is closed
}

func (c *baseClient) String() string {
	return fmt.Sprintf("Redis<%s db:%d>", c.getAddr(), c.cfg.DB)
}

func (c *baseClient) newConn(ctx context.Context) (*pool.Conn, error) {
	cn, err := c.connPool.NewConn(ctx)
	if err != nil {
		return nil, err
	}

	err = c.initConn(ctx, cn)
	if err != nil {
		_ = c.connPool.CloseConn(cn)
		return nil, err
	}

	return cn, nil
}

func (c *baseClient) getConn(ctx context.Context) (*pool.Conn, error) {
	if c.limiter != nil {
		err := c.limiter.Allow()
		if err != nil {
			return nil, err
		}
	}

	cn, err := c._getConn(ctx)
	if err != nil {
		if c.limiter != nil {
			c.limiter.ReportResult(err)
		}
		return nil, err
	}
	return cn, nil
}

func (c *baseClient) _getConn(ctx context.Context) (*pool.Conn, error) {
	cn, err := c.connPool.Get(ctx)
	if err != nil {
		return nil, err
	}

	err = c.initConn(ctx, cn)
	if err != nil {
		c.connPool.Remove(cn)
		return nil, err
	}

	return cn, nil
}

func (c *baseClient) releaseConn(cn *pool.Conn, err error) {
	if c.limiter != nil {
		c.limiter.ReportResult(err)
	}

	if internal.IsBadConn(err, false) {
		c.connPool.Remove(cn)
	} else {
		c.connPool.Put(cn)
	}
}

func (c *baseClient) releaseConnStrict(cn *pool.Conn, err error) {
	if c.limiter != nil {
		c.limiter.ReportResult(err)
	}

	if err == nil || internal.IsRedisError(err) {
		c.connPool.Put(cn)
	} else {
		c.connPool.Remove(cn)
	}
}

func (c *baseClient) initConn(ctx context.Context, cn *pool.Conn) error {
	if cn.Inited {
		return nil
	}
	cn.Inited = true

	if c.cfg.Password == "" &&
		c.cfg.DB == 0 &&
		!c.cfg.readOnly {
		return nil
	}

	conn := newConn(ctx, c.cfg, cn)
	_, err := conn.Pipelined(func(pipe Pipeliner) error {
		if c.cfg.Password != "" {
			pipe.Auth(c.cfg.Password)
		}

		if c.cfg.DB > 0 {
			pipe.Select(c.cfg.DB)
		}

		//if c.cfg.readOnly {
		//	pipe.ReadOnly()
		//}

		return nil
	})
	if err != nil {
		return err
	}

	//if c.cfg.OnConnect != nil {
	//	return c.cfg.OnConnect(conn)
	//}
	return nil
}

func (c *baseClient) process(ctx context.Context, cmd Cmder) error {
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, c.retryBackoff(attempt)); err != nil {
				return err
			}
		}

		cn, err := c.getConn(ctx)
		if err != nil {
			cmd.setErr(err)
			if internal.IsRetryableError(err, true) {
				continue
			}
			return err
		}

		err = cn.WithWriter(ctx, time.Duration(c.cfg.WriteTimeout), func(wr *proto.Writer) error {
			return writeCmd(wr, cmd)
		})
		if err != nil {
			c.releaseConn(cn, err)
			cmd.setErr(err)
			if internal.IsRetryableError(err, true) {
				continue
			}
			return err
		}

		err = cn.WithReader(ctx, c.cmdTimeout(cmd), cmd.readReply)
		c.releaseConn(cn, err)
		if err != nil && internal.IsRetryableError(err, cmd.readTimeout() == nil) {
			continue
		}

		return err
	}

	return cmd.Err()
}

func (c *baseClient) retryBackoff(attempt int) time.Duration {
	return internal.RetryBackoff(attempt, c.cfg.MinRetryBackoff, c.cfg.MaxRetryBackoff)
}

func (c *baseClient) cmdTimeout(cmd Cmder) time.Duration {
	if timeout := cmd.readTimeout(); timeout != nil {
		t := *timeout
		if t == 0 {
			return 0
		}
		return t + 10*time.Second
	}
	return time.Duration(c.cfg.ReadTimeout)
}

// Close closes the client, releasing any open resources.
//
// It is rare to Close a Client, as the Client is meant to be
// long-lived and shared between many goroutines.
func (c *baseClient) Close() error {
	var firstErr error
	if c.onClose != nil {
		if err := c.onClose(); err != nil {
			firstErr = err
		}
	}
	if err := c.connPool.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (c *baseClient) getAddr() string {
	return c.cfg.Addr
}

func (c *baseClient) processPipeline(ctx context.Context, cmds []Cmder) error {
	return c.generalProcessPipeline(ctx, cmds, c.pipelineProcessCmds)
}
type pipelineProcessor func(context.Context, *pool.Conn, []Cmder) (bool, error)

func (c *baseClient) generalProcessPipeline(
	ctx context.Context, cmds []Cmder, p pipelineProcessor,
) error {
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, c.retryBackoff(attempt)); err != nil {
				return err
			}
		}

		cn, err := c.getConn(ctx)
		if err != nil {
			setCmdsErr(cmds, err)
			return err
		}

		canRetry, err := p(ctx, cn, cmds)
		c.releaseConnStrict(cn, err)

		if !canRetry || !internal.IsRetryableError(err, true) {
			break
		}
	}
	return cmdsFirstErr(cmds)
}

func (c *baseClient) pipelineProcessCmds(
	ctx context.Context, cn *pool.Conn, cmds []Cmder,
) (bool, error) {
	err := cn.WithWriter(ctx, time.Duration(c.cfg.WriteTimeout), func(wr *proto.Writer) error {
		return writeCmd(wr, cmds...)
	})
	if err != nil {
		setCmdsErr(cmds, err)
		return true, err
	}

	err = cn.WithReader(ctx, time.Duration(c.cfg.ReadTimeout), func(rd *proto.Reader) error {
		return pipelineReadCmds(rd, cmds)
	})
	return true, err
}

func pipelineReadCmds(rd *proto.Reader, cmds []Cmder) error {
	for _, cmd := range cmds {
		err := cmd.readReply(rd)
		if err != nil && !internal.IsRedisError(err) {
			return err
		}
	}
	return nil
}

type client struct {
	baseClient
	cmdable
}

type Client struct {
	*client
	ctx context.Context
}

// NewClient returns a client to the Redis Server specified by Options.
func NewClient(cfg *Config) *Client {
	cfg.init()

	c := Client{
		client: &client{
			baseClient: baseClient{
				cfg:      cfg,
				connPool: newConnPool(cfg),
			},
		},
		ctx: context.Background(),
	}
	c.init()

	return &c
}

func (c *Client) init() {
	c.cmdable = c.Process
}
func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		panic("nil context")
	}
	clone := *c
	clone.ctx = ctx
	clone.init()
	return &clone
}

// Do creates a Cmd from the args and processes the cmd.
func (c *Client) Do(args ...interface{}) *Cmd {
	return c.DoContext(c.ctx, args...)
}

func (c *Client) DoContext(ctx context.Context, args ...interface{}) *Cmd {
	cmd := NewCmd(args...)
	_ = c.ProcessContext(ctx, cmd)
	return cmd
}

func (c *Client) Process(cmd Cmder) error {
	return c.ProcessContext(c.ctx, cmd)
}

func (c *Client) ProcessContext(ctx context.Context, cmd Cmder) error {
	return c.baseClient.process(ctx, cmd)
}

//------------------------------------------------------------------------------

type conn struct {
	baseClient
	cmdable
	statefulCmdable
}

// Conn is like Client, but its pool contains single connection.
type Conn struct {
	*conn
	ctx context.Context
}

func newConn(ctx context.Context, cfg *Config, cn *pool.Conn) *Conn {
	c := Conn{
		conn: &conn{
			baseClient: baseClient{
				cfg:      cfg,
				connPool: pool.NewSingleConnPool(cn),
			},
		},
		ctx: ctx,
	}
	c.cmdable = c.Process
	c.statefulCmdable = c.Process
	return &c
}

func (c *Conn) Process(cmd Cmder) error {
	return c.ProcessContext(c.ctx, cmd)
}

func (c *Conn) ProcessContext(ctx context.Context, cmd Cmder) error {
	return c.baseClient.process(ctx, cmd)
}

func (c *Conn) Pipelined(fn func(Pipeliner) error) ([]Cmder, error) {
	return c.Pipeline().Pipelined(fn)
}

func (c *Conn) Pipeline() *Pipeline {
	pipe := &Pipeline{
		ctx:  c.ctx,
		exec: c.processPipeline,
	}
	pipe.init()
	return pipe
}