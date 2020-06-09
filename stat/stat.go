/**------------------------------------------------------------**
 * @filename stat/stat.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-25 14:32
 * @desc     go.jd100.com - stat - 统计
 **------------------------------------------------------------**/
package stat

import "go.jd100.com/medusa/stat/prom"

// default stat struct.
var (
	// http
	HTTPClient Stat = prom.HTTPClient
	HTTPServer Stat = prom.HTTPServer
	// storage
	Cache Stat = prom.LibClient
	DB    Stat = prom.LibClient
	// rpc
	RPCClient Stat = prom.RPCClient
	RPCServer Stat = prom.RPCServer
)

// Stat interface.
type Stat interface {
	Timing(name string, time int64, extra ...string)
	Incr(name string, extra ...string) // name,ext...,code
	State(name string, val int64, extra ...string)
}