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
	"log"
	"sync"

	"golang.org/x/net/context"

	"github.com/mitchellh/hashstructure"
)

// TrimSubtrees removes duplicates in the graph
func TrimSubtrees(ctx context.Context, g *Graph) (*Graph, error) {
	lock := new(sync.Mutex)
	values := map[uint64]string{}

	return g.RootFirstTransform(ctx, func(id string, out *Graph) error {
		if id == "root" { // root
			return nil
		}

		lock.Lock()
		defer lock.Unlock()

		value := out.Get(id)
		hash, err := hashstructure.Hash(value, nil)
		if err != nil {
			return err
		}

		// if we haven't seen this value before, register it and return
		target, ok := values[hash]
		if !ok {
			log.Printf("[DEBUG] trim subtrees: registering %q as original\n", id)
			values[hash] = id

			return nil
		}

		log.Printf("[DEBUG] trim subtrees: found duplicate: %q and %q\n", target, id)

		// Point all inbound links to value to target instead
		for _, src := range g.UpEdges(id) {
			log.Printf("[DEBUG] trim subtrees: re-pointing %q from %q to %q\n", src, id, target)
			out.Disconnect(src, id)
			out.Connect(src, target)
		}

		// Remove children and their edges
		for _, child := range g.Descendents(id) {
			log.Printf("[DEBUG] trim subtrees: removing child %q\n", child)
			out.Remove(child)
		}

		// Remove value
		out.Remove(id)

		return nil
	})
}
