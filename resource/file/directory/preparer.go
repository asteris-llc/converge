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
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for Directory
//
// Directory makes sure a directory is present on disk
type Preparer struct {
	// the location on disk to make the directory
	Destination string `hcl:"destination" required:"true" nonempty:"true"`

	// whether or not to create all parent directories on the way up
	CreateAll bool `hcl:"create_all"`
}

// Prepare the new directory
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	return &Directory{
		Destination: p.Destination,
		CreateAll:   p.CreateAll,
	}, nil
}

func init() {
	registry.Register("file.directory", (*Preparer)(nil), (*Directory)(nil))
}
