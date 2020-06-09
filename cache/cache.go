/**------------------------------------------------------------**
 * @filename cache/cache.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-19 11:56
 * @desc     go.jd100.com - cache - 缓存
 **------------------------------------------------------------**/
package cache

import (
	"context"
)

// Conn represents a connection to a Redis server.
type Conn interface {
	// Close closes the connection.
	Close() error
	Err() error
	Do(commandName string, args ...interface{}) (reply interface{}, err error)
	Send(commandName string, args ...interface{}) error
	Flush() error
	Receive() (reply interface{}, err error)
	WithContext(ctx context.Context) Conn
}