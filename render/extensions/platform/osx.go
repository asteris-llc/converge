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
	"log"
	"os/exec"
	"strings"
)

//OSXVers runs /usr/bin/sw_vers to get OSX version information
func (platform *Platform) OSXVers() {
	cmd := "/usr/bin/sw_vers"
	var (
		cmdOut []byte
		err    error
	)

	if cmdOut, err = exec.Command(cmd).Output(); err != nil {
		log.Println("[INFO] Error running %s:%s Will be unable to parse release data", cmd, err)
		return
	}
	platform.ParseOSXVersion(string(cmdOut))
}

//ParseOSXVersion Takes output from /usr/bin/sw_vers and stores in a Platform
func (platform *Platform) ParseOSXVersion(versionData string) {
	lines := strings.Split(versionData, "\n")
	for _, l := range lines {
		s := strings.Split(l, ":")
		switch s[0] {
		case "ProductName":
			platform.Name = strings.Trim(s[1], "\t")
		case "ProductVersion":
			platform.Version = strings.Trim(s[1], "\t")
		case "BuildVersion":
			platform.Build = strings.Trim(s[1], "\t")
		}
	}
}
