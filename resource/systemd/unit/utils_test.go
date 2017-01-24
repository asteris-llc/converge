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

package unit

import (
	"math/rand"
	"strings"
)

// randomizeCase will randomize the character casing in a string
func randomizeCase(s string) string {
	newS := s
	count := rand.Int() % len(s)
	for i := 0; i < count; i++ {
		idx := rand.Int() % len(s)
		up := strings.ToUpper(string(s[idx]))
		newS = replaceAt(newS, up, idx)
	}
	return newS
}

// replaces the character at `at` in `replaceIn` with the string `replaceWith`
func replaceAt(replaceIn, replaceWith string, at int) string {
	if at >= len(replaceIn) {
		return replaceIn
	}
	return replaceIn[0:at] + replaceWith + replaceIn[at+1:]
}
