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

package absent

import (
	"fmt"
	"os"
)

// Content renders a content to disk
type Absent struct {
	Destination string
}

// Check if the content needs to be rendered
func (a *Absent) Check() (status string, willChange bool, err error) {
	_, err = os.Stat(a.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist", a.Destination), false, nil
	} else if err == nil {
		return fmt.Sprintf("%q does exist. will be deleted", a.Destination), true, nil
	} else {
		return "", false, nil
	}
}

// Apply writes the content to disk
func (a *Absent) Apply() (err error) {
	_, err = os.Stat(a.Destination)
	if os.IsNotExist(err) {
		return nil
	} else {
		return os.Remove(a.Destination)
	}
}
