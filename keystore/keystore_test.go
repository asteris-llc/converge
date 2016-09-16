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

package keystore

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreTrustedKey(t *testing.T) {
	ks, ksPath, err := NewTestKeystore()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	defer os.RemoveAll(ksPath)

	output, err := ks.StoreTrustedKey(publicKey)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if filepath.Base(output) != fingerprint {
		t.Errorf("expected finger print %s, got %v", fingerprint, filepath.Base(output))
	}
	if err := ks.DeleteTrustedKey(fingerprint); err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if _, err := os.Lstat(output); !os.IsNotExist(err) {
		t.Errorf("unexpected error %v", err)
	}

	output, err = ks.MaskTrustedSystemKey(fingerprint)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	fileInfo, err := os.Lstat(output)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if fileInfo.Size() != 0 {
		t.Errorf("expected empty file")
	}
}

func TestCheckSignature(t *testing.T) {
	ks, ksPath, err := NewTestKeystore()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	defer os.RemoveAll(ksPath)

	_, err = ks.StoreTrustedKey(publicKey)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	err = ks.CheckSignature(bytes.NewReader(module), bytes.NewReader(signature))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

// NewTestKeystore creates a new KeyStore backed by a temp directory.
func NewTestKeystore() (*Keystore, string, error) {
	dir, err := ioutil.TempDir("", "keystore-test")
	if err != nil {
		return nil, "", err
	}
	localDir := filepath.Join(dir, "trustedkeys")
	userDir := filepath.Join(dir, "/home/converge/trustedkeys")
	systemDir := filepath.Join(dir, "/etc/converge/trustedkeys")
	for _, path := range []string{localDir, userDir, systemDir} {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, "", err
		}
	}
	return New(localDir, userDir, systemDir), dir, nil
}

const fingerprint = "74fdf669f18d59f92b0aaccd720351ff475cc928"

var module = []byte("param \"message\" {\n  default = \"Hello, World!\"\n}\n\nparam \"filename\" {\n  default = \"test.txt\"\n}\n\ntask \"render\" {\n  check = \"cat {{param `filename`}} | tee /dev/stderr | grep -q '{{param `message`}}'\"\n  apply = \"echo '{{param `message`}}' > {{param `filename`}}\"\n}\n")

var signature = []byte(`
-----BEGIN PGP SIGNATURE-----
Version: GnuPG v1

iQEcBAABAgAGBQJXp4W5AAoJEB+82b4TJ9ic970H/jMcsicc2dDoISSJ0JZ8qxVg
gvoqrmbA2Vn0nKpGIL83cf4M6n2f9DWelNNiq5N6Cq/i6+TIUeKFam/4+O7CXmmQ
BhKcxyQg6QDfLG12ivZMyFkKMFkp12NV9qojJwAaDQgeBKZRS4/cKecN/DXuV5H8
SW015eBtPboQZgktftFeq1DtDmj62nKUMMoc0Z0FkXJWGoG+BUk+JHZtFe+D8EDs
yqiIC49EGKvu9rln0gp1iEKKbFzGHwxqERc6a/IbBbRz9wcmcqio0RPcN6DCxEFX
/4zlzOg2GnpiICBPcYpUH7Q08l9OWhk/3DC7evztTnBQav0thH/pfREr590dGJU=
=K1Ms
-----END PGP SIGNATURE-----
`)

var publicKey = []byte(`
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
=6XW+
-----END PGP PUBLIC KEY BLOCK-----`)
