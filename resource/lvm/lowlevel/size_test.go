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

package lowlevel_test

import (
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"

	"testing"
)

// TestParseSize tests ParseSize()
func TestParseSize(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("normal absolute values", func(t *testing.T) {
		size, err := lowlevel.ParseSize("100G")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), size.Size)
		assert.Equal(t, false, size.Relative)
		assert.Equal(t, "G", size.Unit)

		assert.Equal(t, "-L", size.Option())
		assert.Equal(t, "100G", size.String())
	})

	t.Run("normal relative values", func(t *testing.T) {
		size, err := lowlevel.ParseSize("100%FREE")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), size.Size)
		assert.Equal(t, true, size.Relative)
		assert.Equal(t, "%FREE", size.Unit)

		assert.Equal(t, "-l", size.Option())
		assert.Equal(t, "100%FREE", size.String())
	})

	t.Run("percentage units", func(t *testing.T) {
		size, err := lowlevel.ParseSize("99%FREE")
		assert.NoError(t, err)
		assert.Equal(t, "%FREE", size.Unit)

		size, err = lowlevel.ParseSize("99%VG")
		assert.NoError(t, err)
		assert.Equal(t, "%VG", size.Unit)

		size, err = lowlevel.ParseSize("99%PVS")
		assert.NoError(t, err)
		assert.Equal(t, "%PVS", size.Unit)
	})

	t.Run("bad percentage unit", func(t *testing.T) {
		_, err := lowlevel.ParseSize("100%XYZ")
		assert.Error(t, err)
	})

	t.Run("bad percentage (overflow)", func(t *testing.T) {
		_, err := lowlevel.ParseSize("146%FREE")
		assert.Error(t, err)
	})

	t.Run("bad unit", func(t *testing.T) {
		_, err := lowlevel.ParseSize("146X")
		assert.Error(t, err)
	})
}
