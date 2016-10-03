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

package transform

// StringsMapToStringSlice transforms from a map[string]string to a []string. It
// combines the pairs with the provided combination function. If the combination
// function is nil, the pairs will be joined with a space.
func StringsMapToStringSlice(src map[string]string, combine func(string, string) string) (out []string) {
	if combine == nil {
		combine = func(k, v string) string { return k + " " + v }
	}

	for key, val := range src {
		out = append(out, combine(key, val))
	}

	return out
}
