package common

import (
	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/utils"
	"strings"
)

func reverseMap(m map[string]int) map[int]string {
	n := make(map[int]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}

var (
	subjects = []string{
		"语文",
		"数学",
		"英语",
		"物理",
		"化学",
		"生物",
		"历史",
		"地理",
		"政治",
		"文综",
		"理综",
		"其他",
		"科学",
	}

	grades = []string{
		"小一",
		"小二",
		"小三",
		"小四",
		"小五",
		"小六",
		"初一",
		"初二",
		"初三",
		"初四",
		"高一",
		"高二",
		"高三",
	}
	schoolingTypeNameMap = map[string]int{
		"六三": 0,
		"五四": 1,
	}
	schoolingTypeIdMap = reverseMap(schoolingTypeNameMap)
)

const (
	PRIMARY = 0
	JUNIOR = 1
	SENIOR = 2
)

func CheckGrade(name string) bool {
	return utils.Index(grades, name) != -1
}

func GetGradeType(name string) (gradeType int)  {
	if strings.HasPrefix(name, "高") {
		gradeType = SENIOR
	} else if strings.HasPrefix(name, "初") {
		gradeType = JUNIOR
	} else {
		gradeType = PRIMARY
	}
	return
}

func CheckSubject(name string) bool {
	return utils.Index(subjects, name) != -1
}

func SchoolingTypeInt(name string) (id int, err error) {
	id, ok := schoolingTypeNameMap[name]
	if !ok {
		err = errors.ParamsErr
		return
	}
	return
}

func SchoolingTypeStr(id int) (name string, err error) {
	name, ok := schoolingTypeIdMap[id]
	if !ok {
		err = errors.ParamsErr
		return
	}
	return
}
