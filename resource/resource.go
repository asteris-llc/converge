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

// Task does checking as Monitor does, but it can also make changes to make the
// checks pass.
type Task interface {
	Check() (TaskStatus, error)
	Apply() error
}

// Resource adds metadata about the executed tasks
type Resource interface {
	Prepare(Renderer) (Task, error)
}

// Renderer is passed to resources
type Renderer interface {
	Value() (value string, present bool)
	Render(name, content string) (string, error)
}
