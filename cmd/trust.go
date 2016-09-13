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
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/helpers/logging"
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
		defer cancel()
		GracefulExit(cancel)

		// logging
		clog := log.WithField("component", "client")
		ctx = logging.WithLogger(ctx, clog)

		// iterate over key paths
		ks := keystore.Default()
		for _, path := range args {
			url, err := fetch.ResolveInContext(path, "")
			if err != nil {
				clog.WithError(err).Fatal("could not get url")
			}

			ulog := clog.WithField("url", url)

			ulog.Debug("fetching")
			key, err := fetch.Any(ctx, url)
			if err != nil {
				ulog.WithError(err).Fatal("could not retrieve key")
			}

			if viper.GetBool("skip-review") {
				ulog.Warn("skipping fingerprint review")
			} else {
				accepted, err := reviewFingerprint(key)
				if err != nil {
					ulog.WithError(err).Fatal("error reviewing key")
				}
				if !accepted {
					ulog.Warn("key not trusted")
					continue
				}
			}

			keypath, err := ks.StoreTrustedKey(key)
			if err != nil {
				ulog.WithError(err).Fatal("could not add key")
			}

			ulog.Info("stored key at %s", keypath)
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

	RootCmd.AddCommand(trustCmd)
}
