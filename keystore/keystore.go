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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

// A Keystore represents a repository of trusted public keys which can be
// used to verify PGP signatures.
type Keystore struct {
	LocalPath  string
	UserPath   string
	SystemPath string
}

// New returns a new Keystore backed by the provided paths.
func New(localPath, userPath, systemPath string) *Keystore {
	return &Keystore{
		LocalPath:  localPath,
		UserPath:   userPath,
		SystemPath: systemPath,
	}
}

// Default returns a keystore backed by the default local, user, and system paths.
func Default() *Keystore {
	userPath := ""

	usr, err := user.Current()
	if err == nil {
		userPath = filepath.Join(usr.HomeDir, ".converge/trustedkeys")
	}

	return &Keystore{
		LocalPath:  "./trustedkeys",
		UserPath:   userPath,
		SystemPath: "/usr/lib/converge/trustedkeys",
	}
}

// StoreTrustedKey stores the contents of the public key.
func (ks *Keystore) StoreTrustedKey(pubkeyBytes []byte) (string, error) {
	if err := os.MkdirAll(ks.UserPath, 0755); err != nil {
		return "", err
	}

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubkeyBytes))
	if err != nil {
		return "", err
	}

	if len(keyring) < 1 {
		return "", errors.New("cannot store trusted key: empty keyring")
	}

	pubKey := keyring[0].PrimaryKey
	trustedKeyPath := path.Join(ks.UserPath, fmt.Sprintf("%x", pubKey.Fingerprint))
	if err := ioutil.WriteFile(trustedKeyPath, pubkeyBytes, 0644); err != nil {
		return "", err
	}
	return trustedKeyPath, nil
}

// DeleteTrustedKey deletes the trusted key identified by fingerprint.
func (ks *Keystore) DeleteTrustedKey(fingerprint string) error {
	return os.Remove(path.Join(ks.UserPath, fingerprint))
}

// MaskTrustedSystemKey masks the system trusted key identified by fingerprint.
func (ks *Keystore) MaskTrustedSystemKey(fingerprint string) (string, error) {
	dst := path.Join(ks.UserPath, fingerprint)
	return dst, ioutil.WriteFile(dst, []byte(""), 0644)
}

// CheckSignature takes a signed file and a detached signature and verifies if it is signed by a trusted signer.
func (ks *Keystore) CheckSignature(signed, signature io.Reader) error {
	keyring, err := ks.loadKeyring()
	if err != nil {
		return errors.Wrap(err, "error loading keyring")
	}

	signer, err := openpgp.CheckArmoredDetachedSignature(keyring, signed, signature)

	// openpgp has a weird api so we do some custom error handling.
	if err != nil {
		if err == io.EOF {
			return errors.New("no valid signatures found in signature file")
		}

		return err
	}

	if signer == nil {
		return errors.New("invalid signer")
	}

	return nil
}

func (ks *Keystore) loadKeyring() (openpgp.KeyRing, error) {
	var keyring openpgp.EntityList
	trustedKeys := make(map[string]*openpgp.Entity)

	for _, p := range []string{ks.SystemPath, ks.UserPath, ks.UserPath} {
		files, err := ioutil.ReadDir(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, file := range files {
			if file.Size() == 0 {
				delete(trustedKeys, file.Name())
				continue
			}

			trustedKey, err := os.Open(filepath.Join(p, file.Name()))
			if err != nil {
				return nil, err
			}
			defer trustedKey.Close()

			keys, err := openpgp.ReadArmoredKeyRing(trustedKey)
			if err != nil {
				return nil, err
			}
			if len(keys) < 1 {
				return nil, errors.New("empty keyring")
			}

			fingerprint := fmt.Sprintf("%x", keys[0].PrimaryKey.Fingerprint)
			if file.Name() != fingerprint {
				return nil, fmt.Errorf("fingerprint mismatch: %q:%q", file.Name(), fingerprint)
			}

			trustedKeys[fingerprint] = keys[0]
		}
	}

	for _, v := range trustedKeys {
		keyring = append(keyring, v)
	}
	return keyring, nil
}
