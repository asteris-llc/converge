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

package load

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/asteris-llc/converge/resource"
)

// Load a module from a resource. This uses the protocol in the path (or file://
// if not present) to determine from where the module should be loaded.
func Load(source string) (*resource.Module, error) {
	var (
		protocol string
		path     string
	)
	if strings.Contains(source, "://") {
		split := strings.SplitN(source, "://", 2)
		protocol = split[0]
		path = split[1]
	} else {
		protocol = "file"
		path = source
	}

	switch protocol {
	case "file":
		return FromFile(path)

	default:
		return nil, fmt.Errorf("protocol %q is not implemented", protocol)
	}
}

// FromFile loads a module from a file
func FromFile(filename string) (*resource.Module, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{"file", filename}
		}
		return nil, err
	}

	return Parse(content)
}
