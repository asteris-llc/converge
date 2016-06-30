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

package fetch_test

import (
	"testing"

	"github.com/asteris-llc/converge/fetch"
	"github.com/stretchr/testify/assert"
)

func TestResolveInContextBadLoc(t *testing.T) {
	t.Parallel()

	_, err := fetch.ResolveInContext("::", "file://x")
	assert.Error(t, err)
}

func TestResolveInContextBadContext(t *testing.T) {
	t.Parallel()

	_, err := fetch.ResolveInContext("file://x", "::")
	assert.Error(t, err)
}

func TestResolveInContextAbsolute(t *testing.T) {
	t.Parallel()

	resolved, err := fetch.ResolveInContext("/x", "")
	assert.NoError(t, err)
	assert.Equal(t, "file:///x", resolved)
}

func TestResolveInContextRelative(t *testing.T) {
	t.Parallel()

	resolved, err := fetch.ResolveInContext("x", "file:///a/b/c")
	assert.NoError(t, err)
	assert.Equal(t, "file:///a/b/x", resolved)
}

func TestResolveInContextPreservesProtocol(t *testing.T) {
	t.Parallel()

	resolved, err := fetch.ResolveInContext("file://x", "http://a.com/b/c")
	assert.NoError(t, err)
	assert.Equal(t, "file://x", resolved)
}

func TestResolveInContextSelfResolve(t *testing.T) {
	t.Parallel()

	base := "a/b"
	resolved, err := fetch.ResolveInContext(base, base)
	assert.NoError(t, err)
	assert.Equal(t, base, resolved)
}
