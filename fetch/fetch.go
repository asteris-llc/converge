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
	"path"
)

// Any fetches a path based on a
func Any(loc string) ([]byte, error) {
	url, err := parse(loc)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "file":
		return File(path.Join(url.Host, url.Path))
	case "http", "https":
		return HTTP(loc)
	default:
		return nil, fmt.Errorf("protocol %q is not implemented", url.Scheme)
	}
}
