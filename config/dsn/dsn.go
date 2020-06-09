/**------------------------------------------------------------**
 * @filename dsn/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-25 13:38
 * @desc     go.jd100.com - dsn -
 **------------------------------------------------------------**/
package dsn

import (
	"net/url"
	"reflect"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

var _validator *validator.Validate

func init() {
	_validator = validator.New()
}

type DSN struct {
	*url.URL
}

func Parse(rawDSN string) (*DSN, error) {
	u, err := url.Parse(rawDSN)
	return &DSN{URL: u}, err
}

func (d *DSN) Bind(v interface{}) (url.Values, error) {
	assignFuncs := make(map[string]assignFunc)
	if d.User != nil {
		username := d.User.Username()
		password, ok := d.User.Password()
		if ok {
			assignFuncs["password"] = stringsAssignFunc(password)
		}
		assignFuncs["username"] = stringsAssignFunc(username)
	}
	assignFuncs["address"] = addressesAssignFunc(d.Addresses())
	assignFuncs["network"] = stringsAssignFunc(d.Scheme)
	query, err := bindQuery(d.Query(), v, assignFuncs)
	if err != nil {
		return nil, err
	}
	return query, _validator.Struct(v)
}

func addressesAssignFunc(addresses []string) assignFunc {
	return func(v reflect.Value, to tagOpt) error {
		if v.Kind() == reflect.String {
			if addresses[0] == "" && to.Default != "" {
				v.SetString(to.Default)
			} else {
				v.SetString(addresses[0])
			}
			return nil
		}
		if !(v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String) {
			return &BindTypeError{Value: strings.Join(addresses, ","), Type: v.Type()}
		}
		vals := reflect.MakeSlice(v.Type(), len(addresses), len(addresses))
		for i, address := range addresses {
			vals.Index(i).SetString(address)
		}
		if v.CanSet() {
			v.Set(vals)
		}
		return nil
	}
}

func (d *DSN) Addresses() []string {
	switch d.Scheme {
	case "unix", "unixgram", "unixpacket":
		return []string{d.Path}
	}
	return strings.Split(d.Host, ",")
}
