package utils

import "strconv"

func StringToUint(s string) uint {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(val)
}
