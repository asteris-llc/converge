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
	"io/ioutil"
	"log"
	"regexp"
	"text/template"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/extensions"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/parse"
)

type dependencyGenerator func(node *parse.Node) ([]string, error)

var paramSeekerRe = regexp.MustCompile(`\{\{\s*param\s+.(\w+?).\s*\}\}`)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	log.Println("[INFO] resolving dependencies")

	return g.Transform(ctx, func(id string, out *graph.Graph) error {
		if id == "root" { // skip root
			return nil
		}

		node, ok := out.Get(id).(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", out.Get(id))
		}

		depGenerators := []dependencyGenerator{
			getDepends,
			getParams,
		}

		// we have dependencies from various sources, but they're always IDs, so we
		// can connect them pretty easily
		for _, source := range depGenerators {
			deps, err := source(node)
			if err != nil {
				return err
			}
			for _, dep := range deps {

				out.Connect(id, graph.SiblingID(id, dep))
			}
		}
		return nil
	})
}

func getDepends(node *parse.Node) ([]string, error) {
	deps, err := node.GetStringSlice("depends")

	switch err {
	case parse.ErrNotFound:
		return []string{}, nil

	case nil:
		return deps, nil

	default:
		return nil, err
	}
}

func getParams(node *parse.Node) (out []string, err error) {

	// get sibling dependencies. In this case, we need to look for template
	// calls to `param`. Note that I am not proud of this approach. If you,
	// future reader, have a better idea of what to do here: do it!
	//
	// But before you think "oh, I'll just render using a fake param function",
	// remember that every time we add another function in render we'd have to
	// add it here too. If you're reading this, let's have a discussion about
	// what we should do to deduplicate. I'm not sure.
	var strings []string
	strings, err = node.GetStrings()
	if err != nil {
		return nil, err
	}

	type stub struct{}
	language := extensions.MakeLanguage()
	language.On("param", extensions.RegisterExistence(&out, 0))
	language.On("split", extensions.DoNothing())
	for _, s := range strings {
		useless := stub{}
		tmpl, err := template.New("DependencyTemplate").Funcs(language.Funcs).Parse(s)
		if err != nil {
			return out, err
		}
		tmpl.Execute(ioutil.Discard, &useless)

	}
	return out, err
}
