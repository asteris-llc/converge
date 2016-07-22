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
	"regexp"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/parse"
)

var paramSeekerRe = regexp.MustCompile(`\{\{\s*param\s+.(\w+?).\s*\}\}`)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	return g.Transform(func(id string, out *graph.Graph) error {
		select {
		case <-ctx.Done():
			return fmt.Errorf("interrupted at %q", id)
		default:
		}

		if id == "root" { // root
			return nil
		}

		node, ok := out.Get(id).(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", out.Get(id))
		}

		deps, err := node.GetStringSlice("depends")
		if err == nil {
			for _, dep := range deps {
				out.Connect(id, graph.SiblingID(id, dep))
			}
		} else if err != parse.ErrNotFound {
			return err
		}

		// get sibling dependencies. In this case, we need to look for template
		// calls to `param`. Note that I am not proud of this approach. If you,
		// future reader, have a better idea of what to do here: do it!
		//
		// But before you think "oh, I'll just render using a fake param function",
		// remember that every time we add another function in render we'd have to
		// add it here too. If you're reading this, let's have a discussion about
		// what we should do to deduplicate. I'm not sure.
		strings, err := node.GetStrings()
		if err != nil {
			return err
		}

		for _, s := range strings {
			for _, match := range paramSeekerRe.FindAllString(s, -1) {
				name := paramSeekerRe.FindStringSubmatch(match)[1]
				out.Connect(id, graph.SiblingID(id, "param."+name))
			}
		}

		return nil
	})
}
