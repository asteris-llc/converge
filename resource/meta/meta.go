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

package meta

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/asteris-llc/converge/resource"
)

type Meta struct {
	resource.Status

	MetaMap map[string]string

	//Platforms    []map[string]interface{}
}

func (m *Meta) Check(resource.Renderer) (resource.TaskStatus, error) {
	m.Status = resource.Status{Output: []string{m.String()}}

	return m.Status, nil
}

func (m *Meta) Apply() (resource.TaskStatus, error) {
	return m.Status, nil
}

func (m *Meta) String() string {
	// get the fields from the struct, then String them
	stringSlice := []string{"meta:"}

	for field, value := range m.MetaMap {
		stringSlice = append(stringSlice, fmt.Sprintf("%v:\t%v", field, value))

	}
	return strings.Join(stringSlice, "\n\t")
}
