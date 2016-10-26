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
	"path"
	"testing"

	"strings"

	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/helpers/testing/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestHTTP(t *testing.T) {
	// HTTP should load successfully
	defer logging.HideLogs(t)()

	addr, cancel, err := http.ServeFile(path.Join("..", "samples", "basic.hcl"))
	defer cancel()
	require.NoError(t, err)

	_, err = fetch.HTTP(context.Background(), addr)
	assert.NoError(t, err)
}

func TestHTTPNotFound(t *testing.T) {
	// HTTP should not succeed for a bad file
	defer logging.HideLogs(t)()

	addr, cancel, err := http.ServeFile(path.Join("..", "samples", "basic.hcl"))
	defer cancel()
	require.NoError(t, err)

	addr = strings.Replace(addr, "hcl", "nope", 1)

	_, err = fetch.HTTP(context.Background(), addr)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "Fetching "+addr+" failed: 404 Not Found")
	}
}
