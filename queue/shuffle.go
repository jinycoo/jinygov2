/**------------------------------------------------------------**
 * @filename queue/shuffle.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/1 14:06
 * @desc     go.jd100.com - queue - 洗牌
 **------------------------------------------------------------**/
package queue

import "math/rand"

func Shuffle(oarr []int64) []int64 {
	for i := len(oarr) - 1; i >= 0; i-- {
		p := RandInt64(0, int64(i))
		a := oarr[i]
		oarr[i] = oarr[p]
		oarr[p] = a
	}
	return oarr
}

func RandInt64(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}
