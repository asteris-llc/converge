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

package node_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph/node"
	"github.com/stretchr/testify/assert"
)

func BenchmarkWithValue(b *testing.B) {
	source := node.New("test", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.WithValue(i)
	}
}

func TestWithValue(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		fst := node.New("test", 1)
		snd := fst.WithValue(2)

		assert.Equal(t, snd.Value(), 2)
	})

	t.Run("shadowing", func(t *testing.T) {
		// TODO: when this grows any pointers, they should be tested here too
		t.Run("no b to a", func(t *testing.T) {
			fst := node.New("test", 1)
			snd := fst.WithValue(2)
			snd.ID = "other"

			assert.Equal(t, fst.ID, "test")
		})

		t.Run("no b to a", func(t *testing.T) {
			fst := node.New("test", 1)
			snd := fst.WithValue(2)
			fst.ID = "other"

			assert.Equal(t, snd.ID, "test")
		})
	})
}
