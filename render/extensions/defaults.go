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

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// DefaultEnv provides a default implementation for the env function in text
// templates. It operates by determining whether an environment variable
// exists; if so, returns its value, otherwise returns an empty string.
func DefaultEnv(env string) string {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if pair[0] == env {
			return pair[1]
		}
	}
	return ""
}

// DefaultSplit provides a default implementation for the split function in text
// templates. It operates by simply reversing the arguments to split so that it
// works in a reasonable manner when dealing with piped input.
func DefaultSplit(sep, str string) []string {
	return strings.Split(str, sep)
}

// DefaultJoin provides a default implementation for the join function in text
// templates. It operates by simply reversing the arguments to split so that it
// works in a reasonable manner when dealing with piped output.
func DefaultJoin(sep string, src interface{}) (string, error) {
	values := reflect.ValueOf(src)

	if values.Kind() != reflect.Slice {
		container := reflect.MakeSlice(reflect.SliceOf(values.Type()), 1, 1)
		container.Index(0).Set(values)
		values = container
	}

	dest := make([]string, values.Len(), values.Cap())
	for i := 0; i < values.Len(); i++ {
		switch val := values.Index(i).Interface().(type) {
		case string:
			dest[i] = val

		case fmt.Stringer:
			dest[i] = val.String()

		default:
			dest[i] = fmt.Sprintf("%v", val)
		}
	}

	return strings.Join(dest, sep), nil
}

// DefaultJsonify just marshals a value to string
func DefaultJsonify(val interface{}) (string, error) {
	out, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
