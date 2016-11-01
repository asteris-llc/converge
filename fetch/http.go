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

package fetch

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

// HTTP fetches content over HTTP
func HTTP(ctx context.Context, loc string) ([]byte, error) {
	var client http.Client
	req, err := http.NewRequest("GET", loc, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "text/plain")

	req = req.WithContext(ctx)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("Fetching %s failed: %s", loc, response.Status)
	}

	return content, err
}
