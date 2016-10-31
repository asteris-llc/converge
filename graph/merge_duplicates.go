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
	"strings"
	"sync"

	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/mitchellh/hashstructure"
	"golang.org/x/net/context"
)

// SkipMergeFunc will be used to determine whether or not to merge a node
type SkipMergeFunc func(*node.Node) bool

// MergeDuplicates removes duplicates in the graph
func MergeDuplicates(ctx context.Context, g *Graph, skip SkipMergeFunc) (*Graph, error) {
	lock := new(sync.Mutex)
	values := map[uint64]string{}

	logger := logging.GetLogger(ctx).WithField("function", "MergeDuplicates")

	return g.RootFirstTransform(ctx, func(meta *node.Node, out *Graph) error {
		if IsRoot(meta.ID) {
			return nil
		}

		if skip(meta) {
			logger.WithField("id", meta.ID).Debug("skipping by request")
			return nil
		}

		if meta.Value() == nil {
			return nil // not much use in hashing a nil value
		}

		lock.Lock()
		defer lock.Unlock()

		hash, err := hashstructure.Hash(meta.Value(), nil)
		if err != nil {
			return err
		}

		// if we haven't seen this value before, register it and return
		target, ok := values[hash]
		if !ok {
			logger.WithField("id", meta.ID).Debug("registering as original")
			values[hash] = meta.ID

			return nil
		}

		logger.WithField("id", target).WithField("duplicate", meta.ID).Debug("found duplicate")

		// Point all inbound links to value to target instead
		for _, src := range Sources(g.UpEdges(meta.ID)) {
			logger.WithField("src", src).WithField("duplicate", meta.ID).WithField("target", target).Debug("re-pointing dependency")
			out.Disconnect(src, meta.ID)
			out.Connect(src, target)
		}

		// Remove children and their edges
		for _, child := range g.Descendents(meta.ID) {
			logger.WithField("child", child).Debug("removing child")
			out.Remove(child)
		}

		// Remove value
		out.Remove(meta.ID)

		return nil
	})
}

// helper functions for various things we need to merge.
// TODO: find a better home

// SkipModuleAndParams skips trimming modules and params
func SkipModuleAndParams(meta *node.Node) bool {
	base := BaseID(meta.ID)
	return strings.HasPrefix(base, "module") || strings.HasPrefix(base, "param")
}
