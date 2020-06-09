/**------------------------------------------------------------**
 * @filename jiny/server.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-24 09:41
 * @desc     go.jd100.com - jiny - http server
 **------------------------------------------------------------**/
package jiny

import "sync"

var once sync.Once
var defaultEngine *Engine

func engine() *Engine {
	once.Do(func() {
		defaultEngine = Default()
	})
	return defaultEngine
}

// Ping is used to set the general HTTP ping handler.
func Ping(handler HandlerFn) {
	engine().GET("/ping", handler)
}

func Index(handler HandlerFn) {
	engine().GET("/", handler)
}

func Group(relativePath string, handlers ...HandlerFn) *RouterGroup {
	return engine().Group(relativePath, handlers...)
}

func Routes() RoutesInfo {
	return engine().Routes()
}

func Run(addr ...string) (err error) {
	return engine().Run(addr...)
}
