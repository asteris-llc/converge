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

import (
	"bytes"
	"log"

	"golang.org/x/net/context"
)

// Sprinter takes a graph node and stringifies it
type Sprinter func(string, interface{}) (string, error)

// Filter determines whether or not to skip a given node
type Filter func(string, interface{}) bool

// Sprint prints the graph with the given printer and filter
func (g *Graph) Sprint(ctx context.Context, printer Sprinter, filter Filter) (string, error) {
	log.Println("[DEBUG] printing to string")

	var buf bytes.Buffer

	err := g.Walk(
		ctx,
		func(id string, val interface{}) error {
			if !filter(id, val) {
				return nil
			}

			out, err := printer(id, val)
			if err != nil {
				return err
			}

			_, err = buf.WriteString(out)
			if err != nil {
				return err
			}

			return buf.WriteByte('\n')
		},
	)

	return buf.String(), err
}
