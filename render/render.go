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
	"context"
	"fmt"
	"log"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
	"github.com/pkg/errors"
)

// Values for rendering
type Values map[string]interface{}

// Render a graph with the provided values
func Render(ctx context.Context, g *graph.Graph, top Values) (*graph.Graph, error) {
	log.Println("[INFO] rendering")

	if g.Contains("root") {
		g.Add("root", module.NewPreparer(top))
	}
	factory, err := NewFactory(ctx, g)
	if err != nil {
		return nil, err
	}
	return g.RootFirstTransform(ctx, func(id string, out *graph.Graph) error {
		res, ok := out.Get(id).(resource.Resource)
		if !ok {
			return fmt.Errorf("Render only deals with graphs of resource.Resource, node was %T", out.Get(id))
		}
		log.Printf("[DEBUG] render: preparing %q\n", id)
		renderer, err := factory.GetRenderer(id)
		if err != nil {
			return errors.Wrap(err, id)
		}
		rendered, err := res.Prepare(renderer)
		if err != nil {
			return errors.Wrap(err, id)
		}
		out.Add(id, resource.WrapTask(rendered))
		factory.Graph = out
		return nil
	})
}
