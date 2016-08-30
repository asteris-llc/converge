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

// PreNotifyFunc will be called before execution
type PreNotifyFunc func(string) error

// PostNotifyFunc will be called after planning
type PostNotifyFunc func(string, interface{}) error

// NotifyPre will call a function before walking a node
func NotifyPre(pre PreNotifyFunc, inner TransformFunc) TransformFunc {
	return func(id string, g *Graph) error {
		if err := pre(id); err != nil {
			return err
		}

		return inner(id, g)
	}
}

// NotifyPost will call a function after walking a node
func NotifyPost(post PostNotifyFunc, inner TransformFunc) TransformFunc {
	return func(id string, g *Graph) error {
		if err := inner(id, g); err != nil {
			return err
		}

		return post(id, g.Get(id))
	}
}

// Notifier can wrap a graph transform
type Notifier struct {
	Pre  PreNotifyFunc
	Post PostNotifyFunc
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
