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
	"net/url"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for Directory
//
// Directory makes sure a directory is present on disk
type Preparer struct {
	// the location on disk to store the file
	Destination string `hcl:"destination"`
	// the source path to the file
	Source string `hcl:"source"`
	// type of hash for the checksum
	HashType string `hcl:"hash_type"`
	// string hash of the file
	Hash string `hcl:"hash_type"`
}

// Prepare the new directory
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	t := &Fetch{
		Destination: p.Destination,
		Source:      p.Source,
		HashType:    p.HashType,
		Hash:        p.Hash,
	}
	return t, Validate(t)
}

func Validate(t *Fetch) error {
	if t.Destination == "" || t.Source == "" {
		return fmt.Errorf("task requires a `destination` and `source` parameter")
	}
	if _, err := url.Parse(t.Source); err != nil {
		return fmt.Errorf("source paramter is not a valid url or path: %s", err)
	}
	if t.Hash != "" && t.HashType == "" {
		return fmt.Errorf("task requires a `hash_type` parameter when `hash` is given")
	}

	if t.Hash != "" && !isValidHashType(t.HashType) {
		return fmt.Errorf("valid `hash_type` are [md5, sha1, sha256, sha512], recived %q", t.HashType)
	}
	return nil
}

func isValidHashType(str string) bool {
	return str == "md5" || str == "sha1" || str == "sha256" || str == "sha512"
}

func init() {
	registry.Register("file.fetch", (*Preparer)(nil), (*Fetch)(nil))
}
