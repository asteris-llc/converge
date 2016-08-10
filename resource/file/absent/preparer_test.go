// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this absent except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package absent_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/absent"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(absent.Preparer))
}

func TestVaildPreparer(t *testing.T) {
	t.Parallel()
	preparers := []resource.Resource{
		//Test default state
		&absent.Preparer{Destination: "/path/to/absent"},
	}
	errs := []string{
		"",
	}
	helpers.PreparerValidator(t, preparers, errs)
}

func TestInVaildPreparer(t *testing.T) {
	t.Parallel()
	preparers := []resource.Resource{
		//Test useless absent module
		&absent.Preparer{},
	}
	errs := []string{
		"resource requires a `destination` parameter",
	}
	helpers.PreparerValidator(t, preparers, errs)
}
