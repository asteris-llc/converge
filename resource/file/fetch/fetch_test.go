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
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(fetch.Fetch))
}

func TestCheckFileNotExist(t *testing.T) {
	task := fetch.Fetch{
		Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
		Destination: "/tmp/converge.tar.gz",
		HashType:    "md5",
		Hash:        "notarealhashbutstringnonetheless",
	}

	status, err := task.Check(fakerenderer.New())
	assert.NoError(t, err)
	assert.Contains(t, status.Messages(), `"/tmp/converge.tar.gz" does not exist. will be fetched`)
	assert.True(t, status.HasChanges())
}

func TestApplyFileOnDisk(t *testing.T) {

	tmpfile, err := ioutil.TempFile("", "fetch_test.txt")
	ioutil.WriteFile(tmpfile.Name(), []byte("hello"), 0777)
	tmpfile.Close()
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	hash, err := getHash(tmpfile.Name(), "md5")
	assert.NoError(t, err)
	task := fetch.Fetch{
		Source:      tmpfile.Name(),
		Destination: "/tmp/fetch_test2.txt",
		HashType:    "md5",
		Hash:        hash,
	}
	defer os.Remove(task.Source)

	_, err = task.Apply()
	require.NoError(t, err)
	status, err := task.Check(fakerenderer.New())
	assert.NoError(t, err)
	assert.Contains(t, status.Messages(), "Checksums matched")
	assert.False(t, status.HasChanges())
}

func getHash(path, hashType string) (string, error) {
	var h hash.Hash
	switch hashType {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		err := fmt.Errorf("unsupported hashType: %s", hashType)
		return "", err
	}

	f, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("Failed to open file for checksum: %s", err)
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(h, f)
	return string(h.Sum(nil)), err

}
