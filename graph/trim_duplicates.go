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
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/mitchellh/hashstructure"
)

// SkipTrimFunc will be used to determine whether or not to trim a node
type SkipTrimFunc func(string) bool

// TrimDuplicates removes duplicates in the graph
func TrimDuplicates(ctx context.Context, g *Graph, skip SkipTrimFunc) (*Graph, error) {
	lock := new(sync.Mutex)
	values := map[uint64]string{}

	return g.RootFirstTransform(ctx, func(id string, out *Graph) error {
		if id == "root" { // root
			return nil
		}

		if skip(id) {
			log.Printf("[TRACE] trim duplicates: skipping %q by request\n", id)
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
			log.Printf("[TRACE] trim duplicates: registering %q as original\n", id)
			values[hash] = id

			return nil
		}

		log.Printf("[DEBUG] trim duplicates: found duplicate: %q and %q\n", target, id)

		// Point all inbound links to value to target instead
		for _, src := range g.UpEdges(id) {
			log.Printf("[TRACE] trim duplicates: re-pointing %q from %q to %q\n", src, id, target)
			out.Disconnect(src, id)
			out.Connect(src, target)
		}

		// Remove children and their edges
		for _, child := range g.Descendents(id) {
			log.Printf("[TRACE] trim duplicates: removing child %q\n", child)
			out.Remove(child)
		}

		// Remove value
		out.Remove(id)

		return nil
	})
}

// helper functions for various things we need to trim.
// TODO: find a better home

// SkipModuleAndParams skips trimming modules and params
func SkipModuleAndParams(id string) bool {
	base := BaseID(id)
	return strings.HasPrefix(base, "module") || strings.HasPrefix(base, "param")
}
