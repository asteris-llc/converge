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

package extensions

import "strings"

// DefaultSplit provides a default implementation for the split function in text
// templates. It operates by simply reversing the arguments to split so that it
// works in a reasonable manner when dealing with piped input.
func DefaultSplit(sep, str string) []string {
	return strings.Split(str, sep)
}
