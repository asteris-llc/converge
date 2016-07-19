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

package render

import (
	"fmt"
	"log"
	"strings"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
)

type Values map[string]interface{}

func Render(g *graph.Graph, top Values) (*graph.Graph, error) {
	return g.RootFirstTransform(func(id string, out *graph.Graph) error {
		if id == "root" {
			log.Println("[DEBUG] render: wrapping root")
			out.Add(id, module.NewPreparer(top))
		}

		res, ok := out.Get(id).(resource.Resource)
		if !ok {
			return fmt.Errorf("Render only deals with graphs of resource.Resource, node was %T", out.Get(id))
		}

		log.Printf("[DEBUG] render: preparing %q\n", id)

		// determine dot value of the current node - mostly for params and modules
		renderer := &Renderer{Graph: out, ID: id}

		name := graph.BaseID(id)
		if strings.HasPrefix(name, "param") {
			parent, ok := out.GetParent(id).(*module.Module)
			if !ok {
				return fmt.Errorf("Parent of param was not a module, was %T", out.GetParent(id))
			}

			if val, ok := parent.Params[name[len("param."):]]; ok {
				renderer.DotValue = &val
			}
		}

		rendered, err := res.Prepare(renderer)
		if err != nil {
			return err
		}

		out.Add(id, rendered)

		return nil
	})
}
