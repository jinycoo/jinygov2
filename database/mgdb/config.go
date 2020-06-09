/**------------------------------------------------------------**
 * @filename mgdb/config.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/13 14:47
 * @desc     go.jd100.com - mgdb - mongodb config
 **------------------------------------------------------------**/
package mgdb

import (
	"go.jd100.com/medusa/ctime"
)

type Config struct {
	Addr         string // for trace
	DSN          string // write data source name.
	Username     string
	Password     string
	Timeout      ctime.Duration
	Database     string
	IdleTimeout  ctime.Duration // connect max life time.
	QueryTimeout ctime.Duration // query sql timeout
	ExecTimeout  ctime.Duration // execute sql timeout
}
