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

import "testing"

//test output of /usr/sbin/sw_vers
func TestParseOSXSWVers(t *testing.T) {
	string := "ProductName:\tMac OS X\nProductVersion:\t10.11.6\nBuildVersion:\t15G31\n"

	var platform Platform

	platform.ParseOSXVersion(string)

	if platform.Name != "Mac OS X" {
		t.Errorf("ParseOSXVersion Name: wanted Mac OS X, got %q\n", platform.Name)
	}

	if platform.Version != "10.11.6" {
		t.Errorf("ParseOSXVersion Version: wanted 10.11.6, got %q\n", platform.Version)
	}

	if platform.Build != "15G31" {
		t.Errorf("ParseOSXVersion Version: wanted 15G31, got %q\n", platform.Build)
	}

}
