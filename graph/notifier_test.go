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

package graph_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/stretchr/testify/assert"
)

func TestNotifyTransform(t *testing.T) {
	g := graph.New()
	g.Add(node.New("root", 1))

	doNothing := func(string, *graph.Graph) error { return nil }
	returnError := func(string, *graph.Graph) error { return errors.New("error") }

	t.Run("pre", func(t *testing.T) {
		defer logging.HideLogs(t)()

		var ran bool

		notifier := &graph.Notifier{
			Pre: func(string) error {
				ran = true
				return nil
			},
		}

		_, err := g.Transform(
			context.Background(),
			notifier.Transform(doNothing),
		)

		assert.NoError(t, err)
		assert.True(t, ran)
	})

	t.Run("post", func(t *testing.T) {

		var ran bool

		notifier := &graph.Notifier{
			Post: func(string, interface{}) error {
				ran = true
				return nil
			},
		}

		t.Run("no error", func(t *testing.T) {
			defer logging.HideLogs(t)()
			ran = false

			_, err := g.Transform(
				context.Background(),
				notifier.Transform(doNothing),
			)

			assert.NoError(t, err)
			assert.True(t, ran)
		})

		t.Run("error", func(t *testing.T) {
			defer logging.HideLogs(t)()
			ran = false

			_, err := g.Transform(
				context.Background(),
				notifier.Transform(returnError),
			)

			assert.Error(t, err)
			assert.False(t, ran)
		})

	})
}
