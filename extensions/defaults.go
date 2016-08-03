package extensions

import "strings"

func DefaultSplit(sep, str string) []string {
	return strings.Split(str, sep)
}
