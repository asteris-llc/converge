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

package platform

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// LinuxLSB finds Linux LSB files and parses them
// Most modern :inux distributions have standardized on
// /etc/os-release
func (platform *Platform) LinuxLSB() error {
	lsbFile := "/etc/os-release"
	content, err := ioutil.ReadFile(lsbFile)
	if err != nil {
		return errors.Wrapf(err, "%s. Will be unable to parse LSB data", lsbFile)
	}
	platform.ParseLSBContent(string(content))
	return err
}

// ParseLSBContent populates a Platform struct with /etc/os-release data
func (platform *Platform) ParseLSBContent(content string) {
	lines := strings.Split(content, "\n")
	for _, v := range lines {
		s := strings.SplitN(v, "=", 2)
		if len(s) == 2 {
			k, v := s[0], strings.Replace(s[1], "\"", "", -1) //remove quotes
			switch k {
			case "NAME":
				platform.Name = v
			case "ID":
				platform.LinuxDistribution = v
			case "VERSION_ID":
				platform.Version = v
			case "ID_LIKE":
				platform.LinuxLSBLike = strings.Split(v, " ")
			case "BUILD_ID":
				platform.Build = v
			}
		}
	}
}
