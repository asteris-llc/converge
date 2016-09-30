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
	"errors"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func Test_Status_ImplementsCheck(t *testing.T) {
	assert.Implements(t, (*healthcheck.Check)(nil), new(resource.Status))
}

// TestHasChanges exercises all the cases of HasChanges
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

// TestSetError makes sure that the level and error fields are set in all cases.
func TestSetError(t *testing.T) {
	t.Parallel()

	// test the translation of error levels
	fromTo := [][2]resource.StatusLevel{
		{resource.StatusNoChange, resource.StatusFatal},
		{resource.StatusWontChange, resource.StatusFatal},
		{resource.StatusWillChange, resource.StatusCantChange},
		{resource.StatusCantChange, resource.StatusCantChange},
		{resource.StatusFatal, resource.StatusFatal},
	}
	for _, pair := range fromTo {
		testErr := errors.New("test")

		t.Run(pair[0].String(), func(t *testing.T) {
			status := resource.NewStatus()
			status.Level = pair[0]
			status.SetError(testErr)

			assert.Equal(t, pair[1], status.Level, "%s != %s", pair[1], status.Level)
			assert.Equal(t, testErr, status.Error())
		})
	}
}

// TestStatusError makes sure we always return a good error, even if error is
// not set.
func TestStatusError(t *testing.T) {
	t.Parallel()

	// when error is set, just return it
	t.Run("with error", func(t *testing.T) {
		err := errors.New("test")
		status := resource.NewStatus()
		status.SetError(err)

		assert.Equal(t, err, status.Error())
	})

	// otherwise, we'll have to return some generic stuff
	messages := []struct {
		level resource.StatusLevel
		err   error
	}{
		{resource.StatusNoChange, nil},
		{resource.StatusWontChange, nil},
		{resource.StatusWillChange, nil},
		{resource.StatusCantChange, resource.ErrStatusCantChange},
		{resource.StatusFatal, resource.ErrStatusFatal},
	}
	for _, msg := range messages {
		t.Run(msg.level.String(), func(t *testing.T) {
			status := resource.NewStatus()
			status.Level = msg.level
			assert.Equal(t, msg.err, status.Error())
		})
	}
}
