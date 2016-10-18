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

package lock_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/parse/preprocessor/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLock tests that new locks can be created from a node
func TestNewLock(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	stmt := `
task "test" {
	check = "test -f test.txt"
	apply = "touch test.txt"
	lock  = "test.lock"
}
`
	nodes, err := parse.Parse([]byte(stmt))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	lockedNode := nodes[0]
	newLock, err := lock.NewLock(lockedNode)
	require.NoError(t, err)

	assert.NotNil(t, newLock)
	assert.NotNil(t, newLock.LockNode)
	assert.NotNil(t, newLock.UnlockNode)
	assert.NotEmpty(t, newLock.LockID)
	assert.NotEmpty(t, newLock.UnlockID)
}

// TestIsLockNode validates whether a node is a lock node or not
func TestIsLockNode(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("is lock node", func(t *testing.T) {
		stmt := "lock.lock \"test-lock\" {}"
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		assert.True(t, lock.IsLockNode(nodes[0]))
	})

	t.Run("is not lock node", func(t *testing.T) {
		stmt := "task.query \"test-query\" {query=\"hostname\"}"
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		assert.False(t, lock.IsLockNode(nodes[0]))
	})
}

// TestIsUnlockNode validates whether a node is an unlock node or not
func TestIsUnlockNode(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("is unlock node", func(t *testing.T) {
		stmt := "lock.unlock \"test-unlock\" {}"
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		assert.True(t, lock.IsUnlockNode(nodes[0]))
	})

	t.Run("is not lock node", func(t *testing.T) {
		stmt := "task.query \"test-query\" {query=\"hostname\"}"
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		assert.False(t, lock.IsUnlockNode(nodes[0]))
	})
}

// TestNewLockID tests the ability to generate a lock ID from a name
func TestNewLockID(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	name := "test-lock"
	assert.Equal(t, "lock.lock.test-lock", lock.NewLockID(name))
}

// TestNewUnlockID tests the ability to generate an unlock ID from a name
func TestNewUnlockID(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	name := "test-unlock"
	assert.Equal(t, "lock.unlock.test-unlock", lock.NewUnlockID(name))
}

// TestGetLockName tests whether the name of the lock can be derived from a node
// ID
func TestGetLockName(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("lock id", func(t *testing.T) {
		id := "lock.lock.test-lock"
		assert.Equal(t, "test-lock", lock.GetLockName(id))
	})

	t.Run("unlock id", func(t *testing.T) {
		id := "lock.unlock.test-lock"
		assert.Equal(t, "test-lock", lock.GetLockName(id))
	})
}

// TestHasLock tests whether a node with a lock can be detected
func TestHasLock(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("has lock", func(t *testing.T) {
		stmt := `
task "test" {
	check = "test -f test.txt"
	apply = "touch test.txt"
	lock  = "test.lock"
}
`
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		hasLock, err := lock.HasLock(nodes[0])
		require.NoError(t, err)
		assert.True(t, hasLock)
	})

	t.Run("has no lock", func(t *testing.T) {
		stmt := "task.query \"test-query\" {query=\"hostname\"}"
		nodes, err := parse.Parse([]byte(stmt))
		require.NoError(t, err)
		hasLock, err := lock.HasLock(nodes[0])
		require.NoError(t, err)
		assert.False(t, hasLock)
	})
}

// TestGetLockKeyword tests that the correct lock keyword is returned
func TestGetLockKeyword(t *testing.T) {
	assert.Equal(t, "lock.lock", lock.GetLockKeyword())
}

// TestGetUnlockKeyword tests that the correct unlock keyword is returned
func TestGetUnlockKeyword(t *testing.T) {
	assert.Equal(t, "lock.unlock", lock.GetUnlockKeyword())
}
