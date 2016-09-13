package lowlevel

import (
	"fmt"
	"regexp"
	"strconv"
)

// FIXME: should be caseless RE
var pctRE = regexp.MustCompile("^(?i)(\\d+)%(PVS|VG|FREE)$")
var sizeRE = regexp.MustCompile("^(?i)(\\d+)([bskmgtpe])b?$")

func ParseSize(sizeToParse string) (size int64, option string, unit string, err error) {
	if m := pctRE.FindStringSubmatch(sizeToParse); m != nil {
		option = "l"
		unit = "%" + m[2]
		size, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return
		}
		if size > 100 {
			err = fmt.Errorf("size in %% can't be more than 100%%: %d", size)
		}
	} else if m := sizeRE.FindStringSubmatch(sizeToParse); m != nil {
		option = "L"
		unit = m[2]
		size, err = strconv.ParseInt(m[1], 10, 64)
	} else {
		err = fmt.Errorf("size parse error: %s", sizeToParse)
	}
	return
}
