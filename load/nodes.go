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
	"fmt"

	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/graph/node/conditional"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/keystore"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type source struct {
	Parent       string
	ParentName   string
	ParentSource string
	Source       string
}

func (s *source) String() string {
	return fmt.Sprintf("%s (%s)", s.Source, s.Parent)
}

// Nodes loads and parses all resources referred to by the provided url
func Nodes(ctx context.Context, root string, verify bool) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "Nodes")

	toLoad := []*source{{"root", "root", root, root}}

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
			signature, sigErr := fetch.Any(ctx, signatureURL)
			if sigErr != nil {
				return nil, errors.Wrap(sigErr, signatureURL)
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

			newID := graph.ID(current.Parent, resource.ID())
			n := node.New(newID, resource)
			out.Add(n)
			out.ConnectParent(current.Parent, newID)

			n.Source = url
			if current.Source != current.ParentSource {
				n.ParentSource = current.ParentSource
			}

			if resource.IsModule() {
				toLoad = append(
					toLoad,
					&source{
						Parent:       newID,
						ParentName:   resource.ID(),
						ParentSource: url,
						Source:       resource.Source(),
					},
				)
			}
		}
	}
	return out, out.Validate()
}

// expandSwitchMacro is responsible for adding the generated switch nodes into
// the graph.  Nodes inside of the switch macro are added as children to the
// case statements, who are parents of the outer switch statement.  Actual node
// generation happens in parse/preprocessor/switch and we add the nodes into the
// graph here.
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
	switchID := graph.ID(current.Parent, switchNode.ID())
	switchGrNode := node.New(switchID, switchNode)
	g.Add(switchGrNode)
	g.ConnectParent(current.Parent, switchID)

	switchGrNode.AddMetadata(conditional.MetaSwitchName, switchObj.Name)
	switchGrNode.AddMetadata(conditional.MetaType, conditional.NodeCatSwitch)

	var peerList []string
	for _, branch := range switchObj.BranchNames() {
		peerList = append(peerList, "macro.case."+branch)
	}

	for idx, branch := range switchObj.Branches {
		branchNode, err := branch.GenerateNode()
		if err != nil {
			return g, err
		}

		branchID := graph.ID(switchID, branchNode.ID())
		branchGrNode := node.New(branchID, branchNode)
		g.Add(branchGrNode)
		g.ConnectParent(switchID, branchID)

		branchGrNode.AddMetadata(conditional.MetaSwitchName, switchObj.Name)
		branchGrNode.AddMetadata(conditional.MetaUnrenderedPredicate, branch.Predicate)
		branchGrNode.AddMetadata(conditional.MetaBranchName, branch.Name)
		branchGrNode.AddMetadata(conditional.MetaPeers, peerList)
		branchGrNode.AddMetadata(conditional.MetaType, conditional.NodeCatBranch)

		for _, innerNode := range branch.InnerNodes {
			if err := validateInnerNode(innerNode); err != nil {
				return g, err
			}
			innerID := graph.ID(branchID, innerNode.ID())

			condNode := node.New(innerID, innerNode)
			condNode.AddMetadata(conditional.MetaSwitchName, switchObj.Name)
			condNode.AddMetadata(conditional.MetaUnrenderedPredicate, branch.Predicate)
			condNode.AddMetadata(conditional.MetaBranchName, branch.Name)
			condNode.AddMetadata(conditional.MetaPeers, peerList)
			condNode.AddMetadata(conditional.MetaType, conditional.NodeCatResource)

			g.Add(condNode)
			g.ConnectParent(branchID, innerID)
		}
		if idx > 0 {
			parent, _ := switchObj.Branches[idx-1].GenerateNode()

			g.Connect(branchID, graph.ID(switchID, parent.ID()))
		}
	}
	return g, nil
}

// validateInnerNode ensures that we do not nest control statements nor attempt
// to add modules under a switch statement.
func validateInnerNode(node *parse.Node) error {
	switch node.Kind() {
	case "module":
		return errors.New("modules not supported in conditionals")
	case "switch":
		return errors.New("nested conditionals are not supported")
	case "case":
		return errors.New("nested branches are not supported")
	}
	return nil
}
