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

package link

import (
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/resource"
)

// Preparer for Content
type Preparer struct {
	Destination string `hcl:"destination"`
	Source      string `hcl:"source"`
	LinkType    string `hcl:"type"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	source, err := render.Render("source", p.Source)
	if err != nil {
		return nil, err
	}
	ltype, err := render.Render("type", p.LinkType)
	if err != nil {
		return nil, err
	}

	linkTask := &Link{
		Source:      source,
		Destination: destination,
		Type:        LinkType(ltype),
	}
	return linkTask, ValidateTask(linkTask)
}

func ValidateTask(l *Link) error {
	ltype := string(l.Type)
	if ltype != "" && ltype != string(LTSoft) && ltype != string(LTHard) {
		return fmt.Errorf("resource paramter `ltype` can only be %q or %q", LTSoft, LTHard)
	}
	if l.Source == "" || l.Destination == "" {
		return errors.New("resouce `source` or `destination` parameters were empty when attemting to create symbolic link")
	}

	return nil
}
