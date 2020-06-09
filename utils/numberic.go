/**------------------------------------------------------------**
 * @filename utils/numberic.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-06-09 11:19
 * @desc     go.jd100.com - utils - 数字转换
 **------------------------------------------------------------**/
package utils

import (
	"strconv"
	"strings"
)

var (
	// 负数标识
	symbol = [2]string {"负", "負"}
	// 组单位
	groupUnits = map[uint8][2]string {
		0: [2]string{"", ""},
		1: [2]string{"万", "萬"},
		2: [2]string{"亿", "億"},
	}
	// 位
	units = map[uint8][2]string {
		0: [2]string{"", ""},
		1: [2]string{"十", "拾"},
		2: [2]string{"百", "佰"},
		3: [2]string{"千", "仟"},
	}
	// 汉字数值
	numZh = map[uint8][2]string {
		0: [2]string{"零", "零"},
		1: [2]string{"一", "壹"},
		2: [2]string{"二", "贰"},
		3: [2]string{"三", "叁"},
		4: [2]string{"四", "肆"},
		5: [2]string{"五", "伍"},
		6: [2]string{"六", "陆"},
		7: [2]string{"七", "柒"},
		8: [2]string{"八", "捌"},
		9: [2]string{"九", "玖"},
	}
)
/**
 * 获取输入整数转换成汉字表示
 * @param int number 整数
 * @param boolean isTraditional 是否为繁体
 * @return string 汉字数字表示
 */
func ParseNumberTZh(number int, isTraditional bool) (zhNo string) {
	var idx int
	var zh = make([]string, 0)
	if isTraditional {
		idx = 1
	}
	zhNum := number
	if zhNum < 0 {
		zhNum = -zhNum
	}
	zhNo = strconv.Itoa(zhNum)
	list := []byte(zhNo)
	l := len(list)
	m := uint8(l / 4)
	gm := m
	n := uint8(l % 4)
	if n > 0 {
		m += 1
	} else {
		if gm > 0 {
			gm -= 1
		}
	}
	var i,j uint8
	for i = 0; i < m; i++ {
		if  i > 0 || n == 0 {
			n = 4
		}
		for j = 0; j < n; j++ {
			item, _ := strconv.Atoi(string(list[j]))
			unit := units[n-j-1][idx]
			if item == 0 {
				if len(unit) > 0 {
					zh = append(zh, numZh[uint8(item)][idx])
				}
			} else {
				zh = append(zh, numZh[uint8(item)][idx], unit)
			}
		}
		list = list[n:]
		zh = append(zh, groupUnits[gm-i][idx])
	}

	if number < 0 {
		zh = append(zh, symbol[idx])
	}
	return strings.Join(zh, "")
}
