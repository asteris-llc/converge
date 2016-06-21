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
	"errors"
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "serve content over HTTP(S)",
	Aliases: []string{"serve"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check HTTPS
		if viper.GetBool("https") {
			if viper.GetString("certFile") == "" {
				return errors.New("certFile is required for HTTPS")
			}

			if viper.GetString("keyFile") == "" {
				return errors.New("keyFile is required for HTTPS")
			}
		}

		// check module serving
		stat, err := os.Stat(viper.GetString("root"))
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return errors.New("root should be a directory")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		var (
			err error
			s   = server.New(
				ctx,
				viper.GetString("root"),
				viper.GetBool("self-serve"),
			)
		)
		log.Printf("[INFO] serving on %s\n", viper.GetString("addr"))
		if viper.GetBool("https") {
			err = s.ListenAndServeTLS(viper.GetString("addr"), viper.GetString("certFile"), viper.GetString("keyFile"))
		} else {
			err = s.ListenAndServe(viper.GetString("addr"))
		}

		if err != nil {
			log.Fatalf("[FATAL] %s\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// HTTP(S)
	serverCmd.PersistentFlags().String("certFile", "", "certificate file for HTTPS")
	serverCmd.PersistentFlags().String("keyFile", "", "key file for HTTPS")
	serverCmd.PersistentFlags().Bool("https", false, "turn on HTTPS")
	serverCmd.PersistentFlags().StringP("addr", "a", ":8080", "address to listen on")

	// module serving
	serverCmd.PersistentFlags().String("root", ".", "location of modules to serve")

	// self serve
	serverCmd.PersistentFlags().Bool("self-serve", false, "serve own binary for bootstrapping")

	viperBindPFlags(serverCmd.PersistentFlags())
}
