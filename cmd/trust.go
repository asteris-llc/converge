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

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/keystore"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust a key for module verification",
	Long:  `Add keys to the local keystore for use in verifying signed modules.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one key path as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		// iterate over key paths
		ks := keystore.Default()
		for _, path := range args {
			url, err := fetch.ResolveInContext(path, "")
			if err != nil {
				log.Fatalf("[FATAL] %s\n", err)
			}

			log.Printf("[DEBUG] fetching key from %q\n", path)
			key, err := fetch.Any(ctx, url)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not retrieve key: %s\n", url, err)
			}

			if viper.GetBool("skip-review") {
				log.Printf("[WARN] skipping fingerprint review for %q\n", path)
			} else {
				accepted, err := reviewFingerprint(key)
				if err != nil {
					log.Fatalf("error reviewing key: %s\n", err)
				}
				if !accepted {
					log.Printf("[WARN] not trusting %q", path)
					continue
				}
			}

			keypath, err := ks.StoreTrustedKey(key)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not add key: %s\n", url, err)
			}

			log.Printf("[INFO] stored key at %s\n", keypath)
		}
	},
}

func reviewFingerprint(armoredKey []byte) (bool, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(armoredKey))
	if err != nil {
		return false, err
	}

	if len(keyring) < 1 {
		return false, errors.New("keyring is empty")
	}

	key := keyring[0].PrimaryKey
	fmt.Printf("The gpg key fingerprint is %s\n", fmt.Sprintf("%x", key.Fingerprint))

	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Are you sure you want to trust this key (yes/no)? ")

		response, err := in.ReadString('\n')
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "yes":
			return true, nil
		case "no":
			return false, nil
		default:
			fmt.Printf("Please enter 'yes' or 'no'\n")
		}
	}
}

func init() {
	trustCmd.Flags().Bool("skip-review", false, "accept key without fingerprint confirmation")
	viperBindPFlags(trustCmd.Flags())

	RootCmd.AddCommand(trustCmd)
}
