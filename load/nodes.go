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
	"bytes"
	"context"
	"fmt"

	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/keystore"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/pkg/errors"
)

type source struct {
	Parent       string
	ParentSource string
	Source       string
}

func (s *source) String() string {
	return fmt.Sprintf("%s (%s)", s.Source, s.Parent)
}

// Nodes loads and parses all resources referred to by the provided url
func Nodes(ctx context.Context, root string, verify bool) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "Nodes")

	toLoad := []*source{{"root", root, root}}

	out := graph.New()
	out.Add(node.New("root", nil))

	for len(toLoad) > 0 {
		select {
		case <-ctx.Done():
			return nil, errors.New("interrupted")
		default:
		}

		current := toLoad[0]
		toLoad = toLoad[1:]

		url, err := fetch.ResolveInContext(current.Source, current.ParentSource)
		if err != nil {
			return nil, err
		}

		logger.WithField("url", url).Debug("fetching")
		content, err := fetch.Any(ctx, url)
		if err != nil {
			return nil, errors.Wrap(err, url)
		}

		if verify {
			signatureURL := url + ".asc"

			logger.WithField("signatureUrl", signatureURL).Debug("fetching")
			signature, err := fetch.Any(ctx, signatureURL)
			if err != nil {
				return nil, errors.Wrap(err, signatureURL)
			}

			err = keystore.Default().CheckSignature(bytes.NewBuffer(content), bytes.NewBuffer(signature))
			if err != nil {
				return nil, errors.Wrap(err, signatureURL)
			}
		}

		resources, err := parse.Parse(content)
		if err != nil {
			return nil, errors.Wrap(err, url)
		}

		for _, resource := range resources {
			if control.IsSwitchNode(resource) {
				out, err = expandSwitchMacro(content, current, resource, out)
				if err != nil {
					return out, errors.Wrap(err, "unable to load resource")
				}
				continue
			}
			newID := graph.ID(current.Parent, resource.String())
			out.Add(node.New(newID, resource))
			out.ConnectParent(current.Parent, newID)

			if resource.IsModule() {
				toLoad = append(
					toLoad,
					&source{
						Parent:       newID,
						ParentSource: url,
						Source:       resource.Source(),
					},
				)
			}
		}
	}
	return out, out.Validate()
}

func expandSwitchMacro(data []byte, current *source, n *parse.Node, g *graph.Graph) (*graph.Graph, error) {
	if !control.IsSwitchNode(n) {
		return g, nil
	}
	switchObj, err := control.NewSwitch(n, data)
	if err != nil {
		return g, err
	}
	switchNode, err := switchObj.GenerateNode()
	if err != nil {
		return g, err
	}
	switchID := graph.ID(current.Parent, switchNode.String())
	g.Add(switchID, switchNode)
	g.ConnectParent(current.Parent, switchID)
	for _, branch := range switchObj.Branches {
		branchNode, err := branch.GenerateNode()
		if err != nil {
			return g, err
		}
		branchID := graph.ID(switchID, branchNode.String())
		g.Add(branchID, branchNode)
		g.ConnectParent(switchID, branchID)
		for _, innerNode := range branch.InnerNodes {
			if err := validateInnerNode(innerNode); err != nil {
				return g, err
			}
			innerID := graph.ID(branchID, innerNode.String())
			g.Add(innerID, innerNode)
			g.ConnectParent(branchID, innerID)
		}
	}
	return g, nil
}

func validateInnerNode(node *parse.Node) error {
	switch node.Kind() {
	case "module":
		return errors.New("modules not supported in conditionals")
	case "switch":
		return errors.New("nested conditionals are not supported")
	case "case":
		return errors.New("nested branches are not supported")
	}
	fmt.Println("inner node: kind = ", node.Kind())
	return nil
}
