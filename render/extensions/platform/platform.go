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

//Package platform queries the underlying operating system
package platform

import (
	"runtime"
)

//Platform is a struct containing version information for the
//underlying operating system
type Platform struct {
	OS                string
	LinuxDistribution string
	Name              string
	PrettyName        string
	Version           string
	Like              []string
	Build             string
}

//NewPlatform Queries the runtime and then attempts to
//discover version information from the underlying operating system
func NewPlatform() Platform {
	var platform Platform
	platform.OS = runtime.GOOS
	switch platform.OS {
	case "darwin":
		platform.OSXVers()
	case "linux":
		platform.LinuxLSB()
	}
	return platform
}
