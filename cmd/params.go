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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
	"github.com/spf13/pflag"
)

// Once the command line options are parsed, these will hold real values
var paramsJSON string
var params []string

func addParamsArguments(flags *pflag.FlagSet) {
	flags.StringVar(&paramsJSON, "paramsJSON", "{}", "parameters for the top-level module, in JSON format")
	flags.StringSliceVarP(&params, "params", "p", []string{}, "parameters for the top-level module in key=value format")
}

// parseKVPair parses an input of the form "key=value" into its
// corresponding pair of strings. It returns an error on malformed input.
// Everything before the first "=" is considered the key, while everything after
// it is the value.
func parseKVPair(raw string) (string, string, error) {
	pair := strings.SplitN(raw, "=", 2)
	if len(pair) < 2 {
		return "", "", fmt.Errorf("malformed parameter: %v", raw)
	}
	return pair[0], pair[1], nil
}

// insert either puts a key and value into a map, or returns an error if there
// is a duplicate key.
func insert(values resource.Values, key string, value resource.Value) error {
	if _, duplicate := values[key]; duplicate {
		return fmt.Errorf("duplicate entry: found %v=%v and %v=%v", key, values[key], key, value)
	}
	values[key] = resource.Value(value)
	return nil
}

// parseKVPairs parses a list of key=value pairs into a map[string]Value.
func parseKVPairs(pairs []string) (values resource.Values, errors []error) {
	values = make(resource.Values)
	for _, raw := range pairs {
		if key, value, err := parseKVPair(raw); err != nil {
			errors = append(errors, err)
		} else {
			if err = insert(values, key, resource.Value(value)); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return values, errors
}

func getParamsFromFlags(flags *pflag.FlagSet) (vals resource.Values, errors []error) {
	// get parameters passed to the --params flag
	vals, errors = parseKVPairs(params)

	// get parameters passed to the --paramsJSON flag
	jsonParams := resource.Values{}
	if len(paramsJSON) > 0 {
		logrus.Debug("parsing --paramsJSON")

		err := json.Unmarshal([]byte(paramsJSON), &jsonParams)
		// accumulate errors
		if err != nil {
			errors = append(errors, err)
		}
	}

	// merge the two sets of parameters
	for key, value := range jsonParams {
		if err := insert(vals, key, value); err != nil {
			errors = append(errors)
		}
	}

	return vals, errors
}
