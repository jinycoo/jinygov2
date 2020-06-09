/**------------------------------------------------------------**
 * @filename jiny/params.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-01 13:58
 * @desc     go.jd100.com - jiny - params
 **------------------------------------------------------------**/
package jiny

import (
	"math"
	"strconv"
)

func (c *Context) GetParamString(key string, def string) string {
	v := c.Params.ByName(key)
	if v == "" {
		return def
	}

	return v
}

func (c *Context) GetParamInt(key string, def int) int {
	v := c.Params.ByName(key)
	if v == "" {
		return def
	}

	val, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	if val > math.MaxInt32 || val < math.MinInt32 {
		return def
	}
	return val
}

func (c *Context) GetParamInt64(key string, def int64) int64 {
	v := c.Params.ByName(key)
	if v == "" {
		return def
	}

	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	if val > math.MaxInt64 {
		return def
	}

	if val < math.MinInt64 {
		return def
	}

	return val
}

func (c *Context) GetParamUint64(key string, def uint64) uint64 {
	v := c.Params.ByName(key)
	if v == "" {
		return def
	}

	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return def
	}
	if val > math.MaxUint64 {
		return def
	}

	return val
}