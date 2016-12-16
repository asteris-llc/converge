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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// TestPreparerInterface tests that the Preparer interface is properly implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(fetch.Preparer))
}

// TestPreparer tests the valid and invalid cases of Prepare
func TestPreparer(t *testing.T) {
	t.Parallel()

	var (
		fr       = fakerenderer.FakeRenderer{}
		hashType = string(fetch.HashMD5)
		hash     = hex.EncodeToString(md5.New().Sum(nil))
		empty    = ""
		space    = " "
	)

	t.Run("valid", func(t *testing.T) {
		prep := fetch.Preparer{
			Destination: "/tmp/converge.tar.gz",
			Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
			Hash:        &hash,
		}

		t.Run("hashtype", func(t *testing.T) {
			t.Run("md5", func(t *testing.T) {
				prep.HashType = &hashType

				_, err := prep.Prepare(context.Background(), &fr)
				assert.NoError(t, err)
			})

			t.Run("sha1", func(t *testing.T) {
				hashType = string(fetch.HashSHA1)
				prep.HashType = &hashType
				hash = hex.EncodeToString(sha1.New().Sum(nil))
				prep.Hash = &hash

				_, err := prep.Prepare(context.Background(), &fr)
				assert.NoError(t, err)
			})

			t.Run("sha256", func(t *testing.T) {
				hashType = string(fetch.HashSHA256)
				prep.HashType = &hashType
				hash = hex.EncodeToString(sha256.New().Sum(nil))
				prep.Hash = &hash

				_, err := prep.Prepare(context.Background(), &fr)
				assert.NoError(t, err)
			})

			t.Run("sha512", func(t *testing.T) {
				hashType = string(fetch.HashSHA512)
				prep.HashType = &hashType
				hash = hex.EncodeToString(sha512.New().Sum(nil))
				prep.Hash = &hash

				_, err := prep.Prepare(context.Background(), &fr)
				assert.NoError(t, err)
			})
		})

		t.Run("with force", func(t *testing.T) {
			prep.Force = true

			_, err := prep.Prepare(context.Background(), &fr)
			assert.NoError(t, err)
		})
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("source", func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      "",
					HashType:    &hashType,
					Hash:        &hash,
				}
				_, err := prep.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"source\" must contain a value"))
			})

			t.Run("space", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      " ",
					HashType:    &hashType,
					Hash:        &hash,
				}
				_, err := prep.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"source\" must contain a value"))
			})

			t.Run("cannot parse", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      ":test",
					HashType:    &hashType,
					Hash:        &hash,
				}
				_, err := prep.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("failed to parse \"source\": parse %s: missing protocol scheme", prep.Source))
			})
		})

		t.Run("destination", func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "",
					Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
					HashType:    &hashType,
					Hash:        &hash,
				}
				_, err := prep.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"destination\" must contain a value"))
			})

			t.Run("space", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: " ",
					Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
					HashType:    &hashType,
					Hash:        &hash,
				}
				_, err := prep.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"destination\" must contain a value"))
			})
		})

		t.Run("checksum", func(t *testing.T) {
			t.Run("hashtype and hash", func(t *testing.T) {
				t.Run("only hashtype", func(t *testing.T) {
					prep := fetch.Preparer{
						Destination: "/tmp/converge.tar.gz",
						Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
						HashType:    &hashType,
					}
					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, "\"hash\" required with use of \"hash_type\"")
				})

				t.Run("only hash", func(t *testing.T) {
					prep := fetch.Preparer{
						Destination: "/tmp/converge.tar.gz",
						Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
						Hash:        &hash,
					}
					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, "\"hash_type\" required with use of \"hash\"")
				})
			})

			t.Run("hashtype", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
					Hash:        &hash,
				}

				t.Run("empty", func(t *testing.T) {
					prep.HashType = &empty

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash_type\" must be one of \"md5,sha1,sha256,sha512\""))
				})

				t.Run("space", func(t *testing.T) {
					prep.HashType = &space

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash_type\" must be one of \"md5,sha1,sha256,sha512\""))
				})
			})

			t.Run("hash", func(t *testing.T) {
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
					HashType:    &hashType,
				}

				t.Run("empty", func(t *testing.T) {
					prep.Hash = &empty

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" must contain a value"))
				})

				t.Run("space", func(t *testing.T) {
					prep.Hash = &space

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" must contain a value"))
				})
			})

			t.Run("hash length", func(t *testing.T) {
				hash = "invalid"
				prep := fetch.Preparer{
					Destination: "/tmp/converge.tar.gz",
					Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
					Hash:        &hash,
				}

				t.Run("md5", func(t *testing.T) {
					hashType = string(fetch.HashMD5)
					prep.HashType = &hashType

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" is invalid length for %s", hashType))
				})

				t.Run("sha1", func(t *testing.T) {
					hashType = string(fetch.HashSHA1)
					prep.HashType = &hashType

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" is invalid length for %s", hashType))
				})

				t.Run("sha256", func(t *testing.T) {
					hashType = string(fetch.HashSHA256)
					prep.HashType = &hashType

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" is invalid length for %s", hashType))
				})

				t.Run("sha512", func(t *testing.T) {
					hashType = string(fetch.HashSHA512)
					prep.HashType = &hashType

					_, err := prep.Prepare(context.Background(), &fr)
					assert.EqualError(t, err, fmt.Sprintf("\"hash\" is invalid length for %s", hashType))
				})
			})
		})
	})
}
