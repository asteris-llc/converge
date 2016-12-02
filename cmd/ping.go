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
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping a server",
	Long:  "ping a server to check responsiveness",
	Run: func(cmd *cobra.Command, args []string) {
		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		// logging
		plog := log.WithField("component", "client")

		client, err := getInfoClient(ctx, getSecurityConfig())
		if err != nil {
			plog.WithError(err).Fatal("could not get client")
		}

		if err := client.Ping(ctx); err != nil {
			plog.WithError(err).Fatal("could not ping")
		}

		fmt.Println("pong")
	},
}

func init() {
	registerClientSSLFlags(pingCmd.Flags())
	registerRPCFlags(pingCmd.Flags())

	RootCmd.AddCommand(pingCmd)
}
