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

package providers

import (
	"encoding/json"
	"fmt"
	"strings"

	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
)

// RPCProvider is the PrintProvider type for Resources
type RPCProvider struct {
	graphviz.GraphIDProvider
	ShowParams bool
}

// VertexGetID returns the graph ID as the VertexID, possibly maksing it
// depending on the vertext type and configuration.
func (p RPCProvider) VertexGetID(e graphviz.GraphEntity) (pp.VisibleRenderable, error) {
	val, ok := e.Value.(*pb.GraphComponent_Vertex)
	if !ok {
		return pp.VisibleString(e.Name), nil
	}

	switch val.Kind {
	case "param":
		return pp.RenderableString(e.Name, p.ShowParams), nil

	default:
		return pp.VisibleString(e.Name), nil
	}
}

// VertexGetLabel returns a vertext label based on the type of the resource. The
// specific generated labels sare:
//    Templates: Return 'Template' and the file destination
//    Modules: Return 'Module' and the module name
//    Params: Return 'name -> "value"'
//    otherwise: Return 'name'
func (p RPCProvider) VertexGetLabel(e graphviz.GraphEntity) (pp.VisibleRenderable, error) {
	var name string

	if e.Name == rootNodeID {
		name = "/"
	} else {
		name = strings.Split(e.Name, "root/")[1]
	}

	val, ok := e.Value.(*pb.GraphComponent_Vertex)
	if !ok {
		return pp.VisibleString(name), nil
	}

	switch val.Kind {
	case "file.content":
		var dest = new(content.Content)
		if err := json.Unmarshal(val.Details, dest); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal file content")
		}

		return pp.VisibleString(fmt.Sprintf("File: %s", dest.Destination)), nil

	case "module":
		return pp.VisibleString(fmt.Sprintf("Module: %s", name)), nil

	case "param":
		var dest = new(param.Param)
		if err := json.Unmarshal(val.Details, dest); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal param")
		}

		return pp.RenderableString(
			fmt.Sprintf("%s: %s", name, dest.Val),
			p.ShowParams,
		), nil

	default:
		return pp.VisibleString(name), nil
	}
}

// VertexGetProperties sets graphviz attributes based on the type of the
// resource. Specifically, we set the shape to 'component' for Shell preparers
// and 'tab' for templates, and we set the entire root node to be invisible.
func (p RPCProvider) VertexGetProperties(e graphviz.GraphEntity) graphviz.PropertySet {
	properties := make(map[string]string)

	val, ok := e.Value.(*pb.GraphComponent_Vertex)
	if !ok {
		return properties
	}

	switch val.Kind {
	case "task":
		properties["shape"] = "component"

	case "file.content":
		properties["shape"] = "tab"
	}

	return properties
}

// EdgeGetProperties sets attributes for graph edges, specifically making edges
// originating from the Root node invisible.
func (p RPCProvider) EdgeGetProperties(src graphviz.GraphEntity, dst graphviz.GraphEntity) graphviz.PropertySet {
	properties := make(map[string]string)
	return properties
}

// SubgraphMarker identifies the start of subgraphs for resources.
// Specifically, it starts a new subgraph whenever a new 'Module' type resource
// is encountered.
func (p RPCProvider) SubgraphMarker(e graphviz.GraphEntity) graphviz.SubgraphMarkerKey {
	val, ok := e.Value.(*pb.GraphComponent_Vertex)
	if !ok {
		return graphviz.SubgraphMarkerNOP
	}

	if val.Kind == "module" {
		return graphviz.SubgraphMarkerStart
	}

	return graphviz.SubgraphMarkerNOP
}

// NewRPCProvider is a utility function to return a new RPCProvider
func NewRPCProvider() graphviz.PrintProvider {
	return RPCProvider{}
}
