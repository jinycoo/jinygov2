/**------------------------------------------------------------**
 * @filename redis/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-09-18 15:17
 * @desc     go.jd100.com - redis -
 **------------------------------------------------------------**/
package redis

import (
	"go.jd100.com/medusa/container/pool"
)

type Pool struct {
	*pool.Slice
	c *Config
}

//type Config struct {
//	*pool.Config
//
//	// redis name, for trace
//	Name string
//	// The network type, either tcp or unix.
//	// Default is tcp.
//	Proto string
//	// host:port address.
//	Addr string
//	// Optional password. Must match the password specified in the
//	// requirepass server configuration option.
//	Auth string
//	// Dial timeout for establishing new connections.
//	// Default i 5 seconds.
//	DialTimeout ctime.Duration
//	// Timeout for socket reads. If reached, commands will fail
//	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
//	// Default is 3 seconds.
//	ReadTimeout ctime.Duration
//	// Timeout for socket writes. If reached, commands will fail
//	// with a timeout instead of blocking.
//	// Default is ReadTimeout.
//	WriteTimeout ctime.Duration
//}