// Copyright © 2017 Asteris, LLC
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

package unarchive

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Preparer for Unarchive
//
// Unarchive renders unarchive data
type Preparer struct {
	// Source to unarchive - must exist locally
	Source string `hcl:"source" required:"true" nonempty:"true"`

	// Destination for the unarchive - must be a directory
	Destination string `hcl:"destination" required:"true" nonempty:"true"`

	// HashType of the archive. It is the hash function used to generate the
	// checksum hash. Valid types are md5, sha1, sha256, and sha512.
	HashType *string `hcl:"hash_type"`

	// Hash of the archive. It is the checksum hash.
	Hash *string `hcl:"hash" nonempty:"true"`

	// Force indicates whether a file from the unarchived source will replace a
	// file in the destination if it already exists
	// If true, the file will be replaced if:
	// 1. no checksum is provided
	// 2. the checksum of the existing file differs from the checksum provided
	Force bool `hcl:"force"`
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if strings.TrimSpace(p.Source) == "" {
		return nil, errors.New("\"source\" must contain a value")
	}

	_, err := url.Parse(p.Source)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse \"source\"")
	}

	if strings.TrimSpace(p.Destination) == "" {
		return nil, errors.New("\"destination\" must contain a value")
	}

	if p.HashType != nil && p.Hash == nil {
		return nil, errors.New("\"hash\" required with use of \"hash_type\"")
	} else if p.HashType == nil && p.Hash != nil {
		return nil, errors.New("\"hash_type\" required with use of \"hash\"")
	}

	if p.HashType != nil {
		if !isValidHashType(*p.HashType) {
			return nil, fmt.Errorf("\"hash_type\" must be one of \"%s,%s,%s,%s\"", string(HashMD5), string(HashSHA1), string(HashSHA256), string(HashSHA512))
		}
	}

	if p.Hash != nil {
		if strings.TrimSpace(*p.Hash) == "" {
			return nil, errors.New("\"hash\" must contain a value")
		}

		if !isValidHash(*p.HashType, *p.Hash) {
			return nil, fmt.Errorf("\"hash\" is invalid length for %s", *p.HashType)
		}
	}

	unarchive := &Unarchive{
		Source:      p.Source,
		Destination: p.Destination,
		Force:       p.Force,
	}

	if p.HashType != nil {
		unarchive.HashType = *p.HashType
	}

	if p.Hash != nil {
		unarchive.Hash = *p.Hash
	}

	unarchive.fetch = fetch.Fetch{
		Source:      unarchive.Source,
		Destination: os.TempDir(),
		HashType:    unarchive.HashType,
		Hash:        unarchive.Hash,
		Unarchive:   true,
	}

	return unarchive, nil
}

func init() {
	registry.Register("unarchive", (*Preparer)(nil), (*Unarchive)(nil))
}

// isValidHashType returns a bool indicating whether the hash type is valid
func isValidHashType(ht string) bool {
	return ht == string(HashMD5) || ht == string(HashSHA1) || ht == string(HashSHA256) || ht == string(HashSHA512)
}

// isValidHash returns a bool indicating whether the hash length is valid based
// on the hash type specified
func isValidHash(hashType, hash string) bool {
	switch hashType {
	case string(HashMD5):
		return len(hash) == 32
	case string(HashSHA1):
		return len(hash) == 40
	case string(HashSHA256):
		return len(hash) == 64
	case string(HashSHA512):
		return len(hash) == 128
	default:
		return false
	}
}
