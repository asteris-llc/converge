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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	sslUseSSLFlagName   = "use-ssl"
	sslCertFileFlagName = "cert-file"
	sslKeyFileFlagName  = "key-file"
	sslRootCAFlagName   = "ca-file"
)

func registerSSLFlags(flags *pflag.FlagSet) {
	flags.Bool(sslUseSSLFlagName, false, "use SSL for connections")
	flags.String(sslCertFileFlagName, "", "certificate file for SSL")
	flags.String(sslKeyFileFlagName, "", "key file for SSL")
	flags.String(sslRootCAFlagName, "", "CA certificate to trust")
}

func getSSLConfig(serverName string) (*tls.Config, error) {
	if !viper.GetBool(sslUseSSLFlagName) {
		return nil, nil
	}

	config := &tls.Config{
		ServerName: serverName,
	}

	// add root certificates
	if viper.GetString(sslRootCAFlagName) != "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "could not get system cert pool")
		}

		pemBytes, err := ioutil.ReadFile(viper.GetString(sslRootCAFlagName))
		if err != nil {
			return nil, errors.Wrap(err, "could not read specified root CA")
		}

		if added := pool.AppendCertsFromPEM(pemBytes); !added {
			return nil, errors.New("could not append root CA to system roots")
		}

		config.RootCAs = pool
	}

	// add server certificates
	if viper.GetString(sslCertFileFlagName) != "" && viper.GetString(sslKeyFileFlagName) != "" {
		cert, err := tls.LoadX509KeyPair(
			viper.GetString(sslCertFileFlagName),
			viper.GetString(sslKeyFileFlagName),
		)
		if err != nil {
			return nil, err
		}

		config.Certificates = append(config.Certificates, cert)
	}

	return config, nil
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
