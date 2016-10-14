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

import "github.com/asteris-llc/converge/graph/node"

// NotifyFunc will be called before execution
type NotifyFunc func(*node.Node) error

// NotifyPre will call a function before walking a node
func NotifyPre(pre NotifyFunc, inner TransformFunc) TransformFunc {
	return func(meta *node.Node, g *Graph) error {
		if err := pre(meta); err != nil {
			return err
		}

		return inner(meta, g)
	}
}

// NotifyPost will call a function after walking a node
func NotifyPost(post NotifyFunc, inner TransformFunc) TransformFunc {
	return func(meta *node.Node, g *Graph) error {
		if err := inner(meta, g); err != nil {
			return err
		}

		meta, _ = g.Get(meta.ID)

		return post(meta)
	}
}

// Notifier can wrap a graph transform
type Notifier struct {
	Pre  NotifyFunc
	Post NotifyFunc
}

// Transform wraps a TransformFunc with this notifier
func (n *Notifier) Transform(inner TransformFunc) TransformFunc {
	if n == nil {
		return inner
	}

	if n.Pre != nil {
		inner = NotifyPre(n.Pre, inner)
	}

	if n.Post != nil {
		inner = NotifyPost(n.Post, inner)
	}

	return inner
}
