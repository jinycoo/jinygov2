/**------------------------------------------------------------**
 * @filename config/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-15 13:20
 * @desc     go.jd100.com - config -
 **------------------------------------------------------------**/
package config

import (
	"strings"
)

// StringVars []string implement flag.Value
type StringVars []string

func (s StringVars) String() string {
	return strings.Join(s, ",")
}

// Set implement flag.Value
func (s *StringVars) Set(val string) error {
	*s = append(*s, val)
	return nil
}
