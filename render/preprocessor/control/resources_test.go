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

package control_test

import (
	"testing"

	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

// TestInterfacesAreImplemented ensures that all types implement the correct
// interfaces
func TestInterfacesAreImplemented(t *testing.T) {
	t.Run("SwitchPreparer", func(t *testing.T) {
		assert.Implements(t, (*resource.Resource)(nil), new(control.SwitchPreparer))
	})
	t.Run("SwitchTask", func(t *testing.T) {
		assert.Implements(t, (*resource.Task)(nil), new(control.SwitchTask))
	})
	t.Run("CasePreparer", func(t *testing.T) {
		assert.Implements(t, (*resource.Resource)(nil), new(control.CasePreparer))
	})
	t.Run("CaseTask", func(t *testing.T) {
		assert.Implements(t, (*resource.Task)(nil), new(control.CaseTask))
	})
}
