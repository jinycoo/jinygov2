package utils

func Index(array []string, target string) (index int) {
	index = -1
	for index, value :=range array {
		if value == target {
			return index
		}
	}
	return
}