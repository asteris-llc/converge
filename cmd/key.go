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
	"golang.org/x/net/context"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "work with gpg keys",
	Long:  `A suite of commands for working with gpg keys.`,
}

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust a key for module verification",
	Long:  `Add keys to the local keystore for use in verifying signed modules.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need one key path as argument, got 0")
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

		path := args[0]
		ks := keystore.Default()

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
			keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
			if err != nil {
				ulog.WithError(err).Fatal("error reviewing key")
			}

			if len(keyring) < 1 {
				ulog.Fatal("keyring is empty")
			}

			providedFingerprint := viper.GetString("fingerprint")
			fingerprint := fmt.Sprintf("%x", keyring[0].PrimaryKey.Fingerprint)
			accepted := false

			if providedFingerprint == "" {
				accepted = fingerprintPrompt(fingerprint)
			} else {
				accepted = providedFingerprint == fingerprint
			}

			if !accepted {
				ulog.Warn("fingerprint does not match")
				return
			}

		}

		keypath, err := ks.StoreTrustedKey(key)
		if err != nil {
			ulog.WithError(err).Fatal("could not add key")
		}

		ulog.Info("stored key at ", keypath)
	},
}

func fingerprintPrompt(fingerprint string) bool {
	fmt.Printf("The gpg key fingerprint is %s\n", fingerprint)

	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Are you sure you want to trust this key (yes/no)? ")

		response, err := in.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "yes":
			return true
		case "no":
			return false
		default:
			fmt.Printf("Please enter 'yes' or 'no'\n")
		}
	}
}

func init() {
	trustCmd.Flags().Bool("skip-review", false, "accept key without fingerprint confirmation")
	trustCmd.Flags().String("fingerprint", "", "provide a fingerprint instead of prompting")

	keyCmd.AddCommand(trustCmd)
	RootCmd.AddCommand(keyCmd)
}
