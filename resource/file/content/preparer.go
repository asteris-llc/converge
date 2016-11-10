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

package content

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for Content
//
// Content renders content to disk
type Preparer struct {
	// Content is the file content. This will be rendered as a template.
	Content string `hcl:"content"`

	// Destination is the location on disk where the content will be rendered.
	Destination string `hcl:"destination" required:"true" nonempty:"true"`
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	return &Content{
		Destination: p.Destination,
		Content:     p.Content,
	}, nil
}

func init() {
	registry.Register("file.content", (*Preparer)(nil), (*Content)(nil))
}
