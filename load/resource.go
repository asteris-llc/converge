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
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/parse"
	"github.com/hashicorp/hcl"

	// import empty to register types for SetResources
	_ "github.com/asteris-llc/converge/resource/docker/container"
	_ "github.com/asteris-llc/converge/resource/docker/image"
	_ "github.com/asteris-llc/converge/resource/file/content"
	_ "github.com/asteris-llc/converge/resource/file/mode"
	_ "github.com/asteris-llc/converge/resource/module"
	_ "github.com/asteris-llc/converge/resource/param"
	_ "github.com/asteris-llc/converge/resource/shell"
	_ "github.com/asteris-llc/converge/resource/shell/query"
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

		dest, ok := registry.NewByName(node.Kind())
		if !ok {
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
