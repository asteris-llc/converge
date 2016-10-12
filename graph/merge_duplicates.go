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
	"context"
	"strings"
	"sync"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/mitchellh/hashstructure"
)

// SkipMergeFunc will be used to determine whether or not to merge a node
type SkipMergeFunc func(string) bool

// MergeDuplicates removes duplicates in the graph
func MergeDuplicates(ctx context.Context, g *Graph, skip SkipMergeFunc) (*Graph, error) {
	lock := new(sync.Mutex)
	values := map[uint64]string{}

	logger := logging.GetLogger(ctx).WithField("function", "MergeDuplicates")

	return g.RootFirstTransform(ctx, func(id string, out *Graph) error {
		if id == "root" { // root
			return nil
		}

		if skip(id) {
			logger.WithField("id", id).Debug("skipping by request")
			return nil
		}

		lock.Lock()
		defer lock.Unlock()

		value, ok := out.Get(id)
		if !ok {
			return nil // not much use hashing a nil value
		}
		hash, err := hashstructure.Hash(value.Value(), nil)
		if err != nil {
			return err
		}

		// if we haven't seen this value before, register it and return
		target, ok := values[hash]
		if !ok {
			logger.WithField("id", id).Debug("registering as original")
			values[hash] = id

			return nil
		}

		logger.WithField("id", target).WithField("duplicate", id).Debug("found duplicate")

		// Point all inbound links to value to target instead
		for _, src := range Sources(g.UpEdges(id)) {
			logger.WithField("src", src).WithField("duplicate", id).WithField("target", target).Debug("re-pointing dependency")
			out.Disconnect(src, id)
			out.Connect(src, target)
		}

		// Remove children and their edges
		for _, child := range g.Descendents(id) {
			logger.WithField("child", child).Debug("removing child")
			out.Remove(child)
		}

		// Remove value
		out.Remove(id)

		return nil
	})
}

// helper functions for various things we need to merge.
// TODO: find a better home

// SkipModuleAndParams skips trimming modules and params
func SkipModuleAndParams(id string) bool {
	base := BaseID(id)
	return strings.HasPrefix(base, "module") || strings.HasPrefix(base, "param")
}
