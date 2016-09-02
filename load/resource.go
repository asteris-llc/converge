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
	"context"
	"fmt"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker/container"
	"github.com/asteris-llc/converge/resource/docker/image"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/file/group"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/asteris-llc/converge/resource/module"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/hashicorp/hcl"
)

// SetResources loads the resources for each graph node
func SetResources(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "SetResources")
	logger.Info("loading resources")

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

		case "task", "healthcheck.task":
			dest = new(shell.Preparer)

		case "file.content":
			dest = new(content.Preparer)

		case "file.owner":
			dest = new(owner.Preparer)

		case "file.group":
			dest = new(group.Preparer)

		case "file.mode":
			dest = new(mode.Preparer)

		case "docker.image":
			dest = new(image.Preparer)

		case "docker.container":
			dest = new(container.Preparer)

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
