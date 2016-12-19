// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lowlevel

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

// NB: Implement proper case handling as LVM does
//   upper case is S.I. (power of 10)
//   lower case is powers of 1024
//   also special suffixes `S` and `s` exists for sectors
// Related ssue: https://github.com/asteris-llc/converge/issues/448

// Cover values for `66%FREE` and likewise (refer LVM manpages for details).
// See also size_test.go for more usage examples
var pctRE = regexp.MustCompile(`^(?i)(\d+)%(PVS|VG|FREE)$`)

// Cover values for `50G` and likewise (refer LVM manpages for details).
// See also size_test.go for more usage examples.
// Difference between lower/upper cases letters not supported now, see NB above
var sizeRE = regexp.MustCompile(`^(?i)(\d+)([bskmgtpe])b?$`)

// LvmSize represent parsed and validated LVM compatible size
type LvmSize struct {
	Size     int64
	Relative bool
	Unit     string
}

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

// ParseSize parsing and validating sizes in format acceptable by LVM tools
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
			return nil, fmt.Errorf("size in %% can't be more than 100%%: %d", size.Size)
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
