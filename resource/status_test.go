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
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func Test_Status_ImplementsCheck(t *testing.T) {
	assert.Implements(t, (*healthcheck.Check)(nil), new(resource.Status))
}

func TestHasChanges(t *testing.T) {
	t.Parallel()

	// by default, statuses will not change
	t.Run("default", func(t *testing.T) {
		status := new(resource.Status)

		assert.False(t, status.HasChanges())
	})

	// Any diffs imply that this resource will change
	t.Run("diffs", func(t *testing.T) {
		status := new(resource.Status)
		status.AddDifference("x", "a", "b", "c")

		assert.True(t, status.HasChanges())
	})

	// Status level implies changes as well
	caseTable := []struct {
		level      resource.StatusLevel
		willChange bool
	}{
		{resource.StatusNoChange, false},
		{resource.StatusWontChange, false},
		{resource.StatusWillChange, true},
		{resource.StatusFatal, false},
		{resource.StatusCantChange, true},
	}
	for _, row := range caseTable {
		t.Run(
			fmt.Sprintf("status-%s", row.level),
			func(t *testing.T) {
				status := &resource.Status{
					Level: row.level,
				}

				assert.Equal(t, row.willChange, status.HasChanges())
			},
		)
	}
}
