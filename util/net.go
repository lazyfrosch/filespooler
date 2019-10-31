package util

import "strings"

func GetNameFromTCPAddr(address string) string {
	if i := strings.LastIndex(address, ":"); i >= 0 {
		return strings.Trim(address[:i], "[]")
	}
	return address
}
