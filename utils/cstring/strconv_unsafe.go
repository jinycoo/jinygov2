/**------------------------------------------------------------**
 * @filename cstring/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-02 16:15
 * @desc     go.jd100.com - cstring -
 **------------------------------------------------------------**/
package cstring

import "unsafe"

// BytesToString converts byte slice to string.
func UBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes converts string to byte slice.
func UStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}