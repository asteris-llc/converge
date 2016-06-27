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

package resource_test

import (
	"sort"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyTrackerSetDepends(t *testing.T) {
	t.Parallel()

	tracker := new(resource.DependencyTracker)
	tracker.SetDepends([]string{"test"})

	assert.Equal(t, []string{"test"}, *tracker.BaseItems)
}

func TestDependencyTrackerDepends(t *testing.T) {
	t.Parallel()

	tracker := &resource.DependencyTracker{
		BaseItems:     &[]string{"a"},
		ComputedItems: []string{"b"},
	}
	deps := tracker.Depends()
	sort.Strings(deps)

	assert.Equal(
		t,
		[]string{"a", "b"},
		deps,
	)
}

func TestDependencyTrackerDependsDeduplicates(t *testing.T) {
	t.Parallel()

	tracker := &resource.DependencyTracker{
		BaseItems:     &[]string{"a"},
		ComputedItems: []string{"a"},
	}

	assert.Equal(
		t,
		[]string{"a"},
		tracker.Depends(),
	)
}

func TestDependencyTrackerComputeDependencies(t *testing.T) {
	t.Parallel()

	renderer, err := resource.NewRenderer(&resource.Module{
		Resources: []resource.Resource{
			&resource.Param{Name: "test"},
		},
	})
	require.NoError(t, err)

	tracker := new(resource.DependencyTracker)
	err = tracker.ComputeDependencies(
		"test",
		renderer,
		`{{param "test"}}`,
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]string{"param.test"},
		tracker.ComputedItems,
	)
}
