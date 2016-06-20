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

package helpers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

// HTTPServeFile serves a single file on a random port, returning the port it
// chose.
func HTTPServeFile(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	http.HandleFunc(path.Join("/", path.Base(filePath)), func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, path.Base(filePath), time.Now(), f)
	})

	for tries := 5; tries > 0; tries-- {
		// get a random port in the IANA dynamic/ephemeral range
		port := rand.Intn(65535-49151) + 49151

		// start an HTTP server on that port
		errors := make(chan error)
		go func(errors chan error) {
			errors <- http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		}(errors)

		// if it hasn't terminated in .1s, assume it's listening
		dur, _ := time.ParseDuration(".1s")
		select {
		case <-errors:
		case <-time.After(dur):
			return port, nil
		}
	}
	return 0, errors.New("Couldn't find port to listen on")
}
