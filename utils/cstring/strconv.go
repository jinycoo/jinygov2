/**------------------------------------------------------------**
 * @filename cstring/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-02 16:14
 * @desc     go.jd100.com - cstring -
 **------------------------------------------------------------**/
package cstring

import (
	"strconv"
)

func Atoi(b []byte) (int, error) {
	return strconv.Atoi(BytesToString(b))
}

func ParseInt(b []byte, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(BytesToString(b), base, bitSize)
}

func ParseUint(b []byte, base int, bitSize int) (uint64, error) {
	return strconv.ParseUint(BytesToString(b), base, bitSize)
}

func ParseFloat(b []byte, bitSize int) (float64, error) {
	return strconv.ParseFloat(BytesToString(b), bitSize)
}
