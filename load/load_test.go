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

package load_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var samplesDir string

func init() {
	wd, _ := os.Getwd()
	samplesDir = path.Join(wd, "..", "samples")
}

func TestLoadBasic(t *testing.T) {
	defer (helpers.HideLogs(t))()

	_, err := load.Load(path.Join(samplesDir, "basic.hcl"), resource.Values{})
	assert.NoError(t, err)
}

func TestLoadNotExist(t *testing.T) {
	defer (helpers.HideLogs(t))()

	badPath := path.Join(samplesDir, "doesNotExist.hcl")
	_, err := load.Load(badPath, resource.Values{})
	if assert.Error(t, err) {
		assert.EqualError(t, err, fmt.Sprintf("Not found: %q using protocol \"file\"", badPath))
	}
}

func TestLoadFileModule(t *testing.T) {
	defer (helpers.HideLogs(t))()

	_, err := load.Load(path.Join(samplesDir, "sourceFile.hcl"), resource.Values{})
	assert.NoError(t, err)
}

func TestLoadHTTPModule(t *testing.T) {
	defer (helpers.HideLogs(t))()

	addr, cancel, err := helpers.HTTPServeFile(path.Join(samplesDir, "basic.hcl"))
	defer cancel()
	require.NoError(t, err)

	_, err = load.Load(addr, resource.Values{})
	assert.NoError(t, err)
}
