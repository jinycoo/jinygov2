/**------------------------------------------------------------**
 * @filename redis/config.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-19 11:56
 * @desc     go.jd100.com - redis - config
 **------------------------------------------------------------**/
package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.jd100.com/medusa/cache/redis/internal/pool"
	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/errors"
)

type Config struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Proto string
	// host:port address.
	Addr string

	// Dialer creates new network connection and has priority over
	// Network and Addr options.
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error)

	// Hook that is called when new connection is established.
	//OnConnect func(*Conn) error

	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default i 5 seconds.
	DialTimeout ctime.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout ctime.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout ctime.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout ctime.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration

	// Enables read only queries on slave nodes.
	readOnly bool

	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}


func (cfg *Config) init() {
	if cfg.Proto == "" {
		cfg.Proto = "tcp"
	}
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6379"
	}
	if cfg.Dialer == nil {
		cfg.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   time.Duration(cfg.DialTimeout),
				KeepAlive: 5 * time.Minute,
			}
			if cfg.TLSConfig == nil {
				return netDialer.DialContext(ctx, network, addr)
			}
			return tls.DialWithDialer(netDialer, cfg.Proto, cfg.Addr, cfg.TLSConfig)
		}
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10 * runtime.NumCPU()
	}
	if time.Duration(cfg.DialTimeout) == 0 {
		cfg.DialTimeout = ctime.Duration(5 * time.Second)
	}
	switch time.Duration(cfg.ReadTimeout) {
	case -1:
		cfg.ReadTimeout = 0
	case 0:
		cfg.ReadTimeout = ctime.Duration(3 * time.Second)
	}
	switch cfg.WriteTimeout {
	case -1:
		cfg.WriteTimeout = 0
	case 0:
		cfg.WriteTimeout = cfg.ReadTimeout
	}
	if cfg.PoolTimeout == 0 {
		cfg.PoolTimeout = time.Duration(cfg.ReadTimeout) + time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = ctime.Duration(5 * time.Minute)
	}
	if cfg.IdleCheckFrequency == 0 {
		cfg.IdleCheckFrequency = time.Minute
	}

	switch cfg.MinRetryBackoff {
	case -1:
		cfg.MinRetryBackoff = 0
	case 0:
		cfg.MinRetryBackoff = 8 * time.Millisecond
	}
	switch cfg.MaxRetryBackoff {
	case -1:
		cfg.MaxRetryBackoff = 0
	case 0:
		cfg.MaxRetryBackoff = 512 * time.Millisecond
	}
}

// ParseURL parses an URL into Options that can be used to connect to Redis.
func ParseURL(redisURL string) (*Config, error) {
	o := &Config{Proto: "tcp"}
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return nil, errors.New("invalid redis URL scheme: " + u.Scheme)
	}

	if u.User != nil {
		if p, ok := u.User.Password(); ok {
			o.Password = p
		}
	}

	if len(u.Query()) > 0 {
		return nil, errors.New("no options supported")
	}

	h, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		h = u.Host
	}
	if h == "" {
		h = "localhost"
	}
	if p == "" {
		p = "6379"
	}
	o.Addr = net.JoinHostPort(h, p)

	f := strings.FieldsFunc(u.Path, func(r rune) bool {
		return r == '/'
	})
	switch len(f) {
	case 0:
		o.DB = 0
	case 1:
		if o.DB, err = strconv.Atoi(f[0]); err != nil {
			return nil, fmt.Errorf("invalid redis database number: %q", f[0])
		}
	default:
		return nil, errors.New("invalid redis URL path: " + u.Path)
	}

	if u.Scheme == "rediss" {
		o.TLSConfig = &tls.Config{ServerName: h}
	}
	return o, nil
}

func newConnPool(cfg *Config) *pool.ConnPool {
	return pool.NewConnPool(&pool.Options{
		Dialer: func(c context.Context) (net.Conn, error) {
			return cfg.Dialer(c, cfg.Proto, cfg.Addr)
		},
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        time.Duration(cfg.IdleTimeout),
		IdleCheckFrequency: cfg.IdleCheckFrequency,
	})
}
