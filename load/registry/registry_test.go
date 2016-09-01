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

package registry_test

import (
	"testing"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestType struct{}

func TestRegistryRegister(t *testing.T) {
	t.Parallel()

	val := new(TestType)

	t.Run("good", func(t *testing.T) {
		r := registry.New()

		assert.NoError(t, r.Register("test", val))
	})

	t.Run("duplicate", func(t *testing.T) {
		r := registry.New()

		require.NoError(t, r.Register("test", val))
		require.Error(t, r.Register("test", val))
	})
}

func TestRegistryRegisterReverse(t *testing.T) {
	t.Parallel()

	val := new(TestType)

	t.Run("good", func(t *testing.T) {
		r := registry.New()

		assert.NoError(t, r.RegisterReverse(val, "test"))
	})

	t.Run("duplicate", func(t *testing.T) {
		r := registry.New()

		assert.NoError(t, r.RegisterReverse(val, "test"))
		assert.Error(t, r.RegisterReverse(val, "test"))
	})
}

func TestRegistryNewByName(t *testing.T) {
	t.Parallel()

	val := new(TestType)
	r := registry.New()
	require.NoError(t, r.Register("test", val))

	t.Run("good", func(t *testing.T) {
		out, ok := r.NewByName("test")
		assert.IsType(t, (*TestType)(nil), out)
		assert.True(t, ok)
	})

	t.Run("unregistered", func(t *testing.T) {
		out, ok := r.NewByName("unregistered")
		assert.Nil(t, out)
		assert.False(t, ok)
	})
}

func TestRegistryNameForType(t *testing.T) {
	t.Parallel()

	val := new(TestType)
	r := registry.New()
	require.NoError(t, r.Register("test", val))

	t.Run("good", func(t *testing.T) {
		name, ok := r.NameForType(val)
		assert.Equal(t, "test", name)
		assert.True(t, ok)
	})

	t.Run("unregistered", func(t *testing.T) {
		name, ok := r.NameForType(nil)
		assert.Empty(t, name)
		assert.False(t, ok)
	})
}
