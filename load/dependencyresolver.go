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
	"github.com/asteris-llc/converge/parse"
)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(g *graph.Graph) (*graph.Graph, error) {
	return g.Transform(func(id string, val interface{}, edges []string) (interface{}, []string, error) {
		if val == nil && id == "root" { // root
			return val, edges, nil
		}

		node, ok := val.(*parse.Node)
		if !ok {
			return val, edges, fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", val)
		}

		deps, err := node.GetStringSlice("depends")
		if err == nil {
			for _, dep := range deps {
				edges = append(edges, SiblingID(id, dep))
			}
		} else if err != parse.ErrNotFound {
			return val, edges, err
		}

		return val, edges, nil
	})
}
