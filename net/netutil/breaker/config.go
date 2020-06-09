/**------------------------------------------------------------**
 * @filename breaker/config.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-31 14:49
 * @desc     go.jd100.com - breaker - 熔断配置
 **------------------------------------------------------------**/
package breaker

import (
	"time"

	"go.jd100.com/medusa/ctime"
)

type Config struct {
	SwitchOff bool // breaker switch,default off.

	// Hystrix
	Ratio float32
	Sleep ctime.Duration

	// Google
	K float64

	Window  ctime.Duration
	Bucket  int
	Request int64
}

func (conf *Config) fix() {
	if conf.K == 0 {
		conf.K = 1.5
	}
	if conf.Request == 0 {
		conf.Request = 100
	}
	if conf.Ratio == 0 {
		conf.Ratio = 0.5 // default request half open
	}
	if conf.Sleep == 0 { // default sleep 500ms
		conf.Sleep = ctime.Duration(500 * time.Millisecond)
	}
	if conf.Bucket == 0 { // default 10
		conf.Bucket = 10
	}
	if conf.Window == 0 { // default 3s
		conf.Window = ctime.Duration(3 * time.Second)
	}
}
