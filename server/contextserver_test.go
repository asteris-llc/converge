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

package server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testServeFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

var mux *http.ServeMux

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/ping", testServeFunc)
}

func TestContextServerListenAndServe(t *testing.T) {
	defer (helpers.HideLogs(t))()

	ctx, cancel := context.WithCancel(context.Background())

	server := server.NewContextServer(ctx)
	go func() {
		err := server.ListenAndServe("localhost:18080", mux)
		assert.NoError(t, err)
	}()

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:18080/ping")
	require.NoError(t, err)

	val, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(val))

	cancel()
	time.Sleep(500 * time.Millisecond)
}
