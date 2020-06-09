/**------------------------------------------------------------**
 * @filename pool/pool.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-09-18 15:25
 * @desc     go.jd100.com - pool - pool
 **------------------------------------------------------------**/
package pool

import (
	"context"
	"io"
	"time"

	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/errors"
)

var (
	// ErrPoolExhausted connections are exhausted.
	ErrPoolExhausted = errors.New("container/pool exhausted")
	// ErrPoolClosed connection pool is closed.
	ErrPoolClosed = errors.New("container/pool closed")

	// nowFunc returns the current time; it's overridden in tests.
	nowFunc = time.Now
)

// Config is the pool configuration struct.
type Config struct {
	// Active number of items allocated by the pool at a given time.
	// When zero, there is no limit on the number of items in the pool.
	Active int
	// Idle number of idle items in the pool.
	Idle int
	// Close items after remaining item for this duration. If the value
	// is zero, then item items are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout ctime.Duration
	// If WaitTimeout is set and the pool is at the Active limit, then Get() waits WatiTimeout
	// until a item to be returned to the pool before returning.
	WaitTimeout ctime.Duration
	// If WaitTimeout is not set, then Wait effects.
	// if Wait is set true, then wait until ctx timeout, or default flase and return directly.
	Wait bool
}

type item struct {
	createdAt time.Time
	c         io.Closer
}

func (i *item) expired(timeout time.Duration) bool {
	if timeout <= 0 {
		return false
	}
	return i.createdAt.Add(timeout).Before(nowFunc())
}

func (i *item) close() error {
	return i.c.Close()
}

// Pool interface.
type Pool interface {
	Get(ctx context.Context) (io.Closer, error)
	Put(ctx context.Context, c io.Closer, forceClose bool) error
	Close() error
}
