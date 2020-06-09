/**------------------------------------------------------------**
 * @filename string/format.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-25 16:55
 * @desc     go.jd100.com - string - format
 **------------------------------------------------------------**/
package cstring

import "fmt"

func Sprint(template string, args ...interface{}) (message string) {
	message = template
	if message == "" && len(args) > 0 {
		message = fmt.Sprint(args...)
	} else if message != "" && len(args) > 0 {
		message = fmt.Sprintf(template, args...)
	}
	return
}