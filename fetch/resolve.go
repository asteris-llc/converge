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
	"net/url"
	"path"
)

// ResolveInContext resolves a path relative to another
func ResolveInContext(loc, context string) (string, error) {
	url, err := parse(loc)
	if err != nil {
		return "", err
	}

	base, err := parse(context)
	if err != nil {
		return "", err
	}

	if !path.IsAbs(url.Path) && (url.Scheme == "" || url.Scheme == base.Scheme) {
		path := path.Join(path.Dir(base.Path), url.Path)
		*url = *base // shallow copy of the rest of the fields
		url.Path = path
	}

	return url.String(), nil
}

func parse(source string) (*url.URL, error) {
	url, err := url.Parse(source)
	if err != nil {
		return url, err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	return url, nil
}
