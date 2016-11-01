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
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(fetch.Preparer))
}

func TestVaildPreparer(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := fetch.Preparer{
		Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
		Destination: "/tmp/converge.tar.gz",
		HashType:    "md5",
		Hash:        "notarealhashbutstringnonetheless",
	}
	_, err := prep.Prepare(&fr)
	assert.NoError(t, err)
}

func TestInVaildPreparerNoSourceOrDestination(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := fetch.Preparer{
		Destination: "/tmp/converge.tar.gz",
		HashType:    "md5",
		Hash:        "notarealhashbutstringnonetheless",
	}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a `destination` and `source` parameter"))
	prep = fetch.Preparer{
		HashType: "md5",
		Hash:     "notarealhashbutstringnonetheless",
	}
	_, err = prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a `destination` and `source` parameter"))
	prep = fetch.Preparer{
		Source:   "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
		HashType: "md5",
		Hash:     "notarealhashbutstringnonetheless",
	}
	_, err = prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a `destination` and `source` parameter"))
}

func TestInVaildPreparerNoHashType(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := fetch.Preparer{
		Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
		Destination: "/tmp/converge.tar.gz",
		Hash:        "notarealhashbutstringnonetheless",
	}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a `hash_type` parameter when `hash` is given"))
}

func TestInVaildPreparerBadHashType(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := fetch.Preparer{
		Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
		Destination: "/tmp/converge.tar.gz",
		Hash:        "notarealhashbutstringnonetheless",
		HashType:    "bad",
	}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, "valid `hash_type` are [md5, sha1, sha256, sha512], recived \"bad\"")
}
