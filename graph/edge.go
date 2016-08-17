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

package graph

import "github.com/hashicorp/terraform/dag"

// ParentEdge marks an edge as signifying a parent/child relationship
type ParentEdge struct {
	dag.Edge
}

// NewParentEdge constructs a new ParentEdge between the given vertices
func NewParentEdge(parent, child string) *ParentEdge {
	return &ParentEdge{Edge: dag.BasicEdge(parent, child)}
}

// Sources gets the sources from slice of edges
func Sources(edges []dag.Edge) (sources []string) {
	for _, edge := range edges {
		sources = append(sources, edge.Source().(string))
	}

	return sources
}

// Targets gets the targets from slice of edges
func Targets(edges []dag.Edge) (targets []string) {
	for _, edge := range edges {
		targets = append(targets, edge.Target().(string))
	}

	return targets
}
