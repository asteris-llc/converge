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

package directory

import (
	"strconv"

	"github.com/asteris-llc/converge/resource"
)

// Preparer for Directory
//
// Directory makes sure a directory is present on disk
type Preparer struct {
	// the location on disk to make the directory
	Destination string `hcl:"destination"`

	// whether or not to create all parent directories on the way up
	CreateAll string `hcl:"create_all" doc_type:"bool"`
}

// Prepare the new directory
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}

	createAllRaw, err := render.Render("create_all", p.CreateAll)
	if err != nil {
		return nil, err
	}
	createAll, err := strconv.ParseBool(createAllRaw)
	if err != nil {
		return nil, err
	}

	return &Directory{
		Destination: destination,
		CreateAll:   createAll,
	}, nil
}
