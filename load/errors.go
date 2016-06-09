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
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl/hcl/token"
)

// ParseError is returned for errors in parsing the AST into a config.
type ParseError struct {
	Pos     token.Pos
	Message string
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("At %s: %s", err.Pos, err.Message)
}

// MultiError combines multiple errors into one
type MultiError []error

func (err MultiError) Error() string {
	max := len(err) - 1
	var b bytes.Buffer
	for i, e := range err {
		b.WriteString(e.Error())
		if i != max {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// NotFoundError indicates an error in the loading process
type NotFoundError struct {
	Protocol string
	Path     string
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("Not found: %q using protocol %q", err.Path, err.Protocol)
}
