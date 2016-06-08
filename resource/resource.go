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

package resource

import (
	"fmt"
	"regexp"
)

// Monitor checks if a resource is correct.
type Monitor interface {
	Check() (string, bool, error)
}

// Task does checking as Monitor does, but it can also make changes to make the
// checks pass.
type Task interface {
	Monitor
	Apply() (string, bool, error)
}

// Resource adds metadata about the executed tasks
type Resource interface {
	Name() string
	Prepare(*Module) error
	Validate() error
	Depends() []string
	SetDepends([]string)
}

// Parent expresses a resource that has sub-resources instead of being
// executable
type Parent interface {
	Children() []Resource
}

//LongName returns the name of the resource wrapped around its type
//e.g template(a)
func LongName(res Resource) string {
	switch res.(type) {
	case *Module:
		return fmt.Sprintf("module(%s)", res.Name())
	case *ModuleTask:
		return fmt.Sprintf("module_task(%s)", res.Name())
	case *Template:
		return fmt.Sprintf("template(%s)", res.Name())
	case *Param:
		return fmt.Sprintf("param(%s)", res.Name())
	case Task:
		return fmt.Sprintf("task(%s)", res.Name())
	case Monitor:
		return fmt.Sprintf("monitor(%s)", res.Name())
	}
	return fmt.Sprintf("unsupported(%s)", res.Name())
}

//ShortName returns the name of a Resource
//If longName = "type(a)" it returns "a" otherwise it
//returns longName
var shortNameRegex = regexp.MustCompile(`\w+\((\w+)\)`)

func ShortName(longName string) string {
	res := shortNameRegex.FindStringSubmatch(longName)
	if res == nil || res[1] == "" {
		return longName
	}
	return res[1]
}
