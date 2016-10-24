package lowlevel

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

// FIXME: should be caseless RE
var pctRE = regexp.MustCompile("^(?i)(\\d+)%(PVS|VG|FREE)$")
var sizeRE = regexp.MustCompile("^(?i)(\\d+)([bskmgtpe])b?$")

// LvmSize represent parsed and validated LVM compatible size
type LvmSize struct {
	Size     int64
	Relative bool
	Unit     string
}

// FIXME: add accessors and unpublish fields to make it immutable?

// String reconstruct size to LVM compatible form
func (size *LvmSize) String() string {
	return fmt.Sprintf("%d%s", size.Size, size.Unit)
}

// Option reconstruct -l/-L parameter from size.Relative
func (size *LvmSize) Option() string {
	if size.Relative {
		return "-l"
	}
	return "-L"
}

// CommandLine return part of command line for calling LVM tools like lvcreate
func (size *LvmSize) CommandLine() [2]string {
	s := size.String()
	o := size.Option()
	return [2]string{o, s}
}

func ParseSize(sizeToParse string) (*LvmSize, error) {
	var err error
	size := &LvmSize{}
	if m := pctRE.FindStringSubmatch(sizeToParse); m != nil {
		size.Relative = true
		size.Unit = "%" + m[2]
		size.Size, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "Parse LVM size")
		}
		if size.Size > 100 {
			return nil, fmt.Errorf("size in %% can't be more than 100%%: %d", size)
		}
	} else if m := sizeRE.FindStringSubmatch(sizeToParse); m != nil {
		size.Relative = false
		size.Unit = m[2]
		size.Size, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("size parse error: %s", sizeToParse)
	}
	return size, nil
}
