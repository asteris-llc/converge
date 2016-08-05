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
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/absent"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/asteris-llc/converge/resource/file/file"
	"github.com/asteris-llc/converge/resource/file/link"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/asteris-llc/converge/resource/file/touch"
	"github.com/asteris-llc/converge/resource/module"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/hashicorp/hcl"
)

// SetResources loads the resources for each graph node
func SetResources(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	log.Println("[INFO] loading resources")

	return g.Transform(ctx, func(id string, out *graph.Graph) error {
		if id == "root" { // root
			return nil
		}

		node, ok := out.Get(id).(*parse.Node)
		if !ok {
			return fmt.Errorf("SetResources can only be used on Graphs of *parse.Node. I got %T", out.Get(id))
		}

		var dest resource.Resource
		switch node.Kind() {
		case "param":
			dest = new(param.Preparer)

		case "module":
			dest = new(module.Preparer)

		case "task":
			dest = new(shell.Preparer)

		case "file":
			dest = new(file.Preparer)

		case "file.content":
			dest = new(content.Preparer)
		case "file.file":
			dest = new(file.Preparer)
		case "file.absent":
			dest = new(absent.Preparer)
		case "file.link":
			dest = new(link.Preparer)
		case "file.directory":
			dest = new(directory.Preparer)
		case "file.touch":
			dest = new(touch.Preparer)
		case "file.mode":
			dest = new(mode.Preparer)
		case "file.owner":
			dest = new(owner.Preparer)

		default:
			return fmt.Errorf("%q is not a valid resource type in %q", node.Kind(), node)
		}

		err := hcl.DecodeObject(dest, node.ObjectItem)
		if err != nil {
			return err
		}

		out.Add(id, dest)

		return nil
	})
}
