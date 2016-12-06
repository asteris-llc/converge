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
	"github.com/asteris-llc/converge/rpc"
	"github.com/fgrid/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	rpcNoTokenFlagName  = "no-token"
	rpcTokenFlagName    = "rpc-token"
	sslCertFileFlagName = "cert-file"
	sslKeyFileFlagName  = "key-file"
	sslCAFlagName       = "ca-file"
	sslUseSSLFlagName   = "use-ssl"
)

func registerSSLFlags(flags *pflag.FlagSet) {
	flags.Bool(sslUseSSLFlagName, false, "use SSL for connections")
	flags.String(sslCertFileFlagName, "", "certificate file for SSL")
	flags.String(sslKeyFileFlagName, "", "key file for SSL")
	flags.String(sslCAFlagName, "", "CA certificate to trust")
}

func registerClientSSLFlags(flags *pflag.FlagSet) {
	flags.Bool(sslUseSSLFlagName, false, "use SSL for connections")
	flags.String(sslCAFlagName, "", "CA certificate to trust")
}

func getSecurityConfig() *rpc.Security {
	out := &rpc.Security{
		Token:  getToken(),
		UseSSL: usingSSL(),
	}

	if usingSSL() {
		out.CertFile = getCertFileLoc()
		out.KeyFile = getKeyFileLoc()
		out.CAFile = getCAFileLoc()
	}

	return out
}

func validateSSL() error {
	if !usingSSL() {
		return nil
	}

	if getCertFileLoc() == "" {
		return fmt.Errorf("%s is required for SSL usage", sslCertFileFlagName)
	}

	if getKeyFileLoc() == "" {
		return fmt.Errorf("%s is required for SSL usage", sslKeyFileFlagName)
	}

	return nil
}

func usingSSL() bool         { return viper.GetBool(sslUseSSLFlagName) }
func getCertFileLoc() string { return viper.GetString(sslCertFileFlagName) }
func getKeyFileLoc() string  { return viper.GetString(sslKeyFileFlagName) }
func getCAFileLoc() string   { return viper.GetString(sslCAFlagName) }

// Token

func getToken() string { return viper.GetString(rpcTokenFlagName) }

func maybeSetToken() {
	if viper.GetBool(rpcNoTokenFlagName) {
		log.Warning("no token set, server is unauthenticated. This should *only* be used for development.")
		return
	}

	if getToken() == "" && getLocal() {
		viper.Set(rpcTokenFlagName, uuid.NewV4().String())
		log.WithField("token", getToken()).Warn("setting session-local token")
	}
}
