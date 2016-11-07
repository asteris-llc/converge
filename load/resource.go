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

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/hcl"

	// import empty to register types for SetResources
	_ "github.com/asteris-llc/converge/resource/docker/container"
	_ "github.com/asteris-llc/converge/resource/docker/image"
	_ "github.com/asteris-llc/converge/resource/docker/volume"
	_ "github.com/asteris-llc/converge/resource/file/content"
	_ "github.com/asteris-llc/converge/resource/file/directory"
	_ "github.com/asteris-llc/converge/resource/file/mode"
	_ "github.com/asteris-llc/converge/resource/group"
	_ "github.com/asteris-llc/converge/resource/module"
	_ "github.com/asteris-llc/converge/resource/package/apt"
	_ "github.com/asteris-llc/converge/resource/package/rpm"
	_ "github.com/asteris-llc/converge/resource/param"
	_ "github.com/asteris-llc/converge/resource/shell"
	_ "github.com/asteris-llc/converge/resource/shell/query"
	_ "github.com/asteris-llc/converge/resource/user"
	_ "github.com/asteris-llc/converge/resource/wait"
	_ "github.com/asteris-llc/converge/resource/wait/port"
	"golang.org/x/net/context"
)

// SetResources loads the resources for each graph node
func SetResources(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "SetResources")
	logger.Debug("loading resources")

	return g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		if graph.IsRoot(meta.ID) {
			return nil
		}

		raw, ok := meta.Value().(*parse.Node)
		if !ok {
			return fmt.Errorf("SetResources can only be used on Graphs of *parse.Node. I got %T", meta.Value())
		}

		dest, ok := registry.NewByName(raw.Kind())
		if !ok {
			return fmt.Errorf("%q is not a valid resource type in %q", raw.Kind(), raw)
		}

		res, ok := dest.(resource.Resource)
		if !ok {
			return fmt.Errorf("%q is not a valid resource, got %T", raw.Kind(), dest)
		}

		preparer := resource.NewPreparer(res)

		err := hcl.DecodeObject(&preparer.Source, raw.ObjectItem.Val)
		if err != nil {
			return err
		}

		out.Add(meta.WithValue(preparer))
		return nil
	})
}
