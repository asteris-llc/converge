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
	"testing"

	"github.com/asteris-llc/converge/render"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// set up a FlagSet for testing
func setupFlags(params, paramsJSON string) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("TestGetParamsFromFlags", pflag.PanicOnError)
	registerParamsFlags(flagSet)
	// mirror actual usage by using Parse rather than Set
	cmdline := []string{"apply"}
	if params != "" {
		cmdline = append(cmdline, "--params", params)
	}
	if paramsJSON != "" {
		cmdline = append(cmdline, "--paramsJSON", paramsJSON)
	}

	if err := flagSet.Parse(append(cmdline, "samples/test.hcl")); err != nil {
		panic(err)
	}

	return flagSet
}

func TestRegisterParamsFlags(t *testing.T) {
	flagSet := pflag.NewFlagSet("", pflag.PanicOnError)
	registerParamsFlags(flagSet)
	assert.True(t, flagSet.HasAvailableFlags())
}

func TestParseKVPairFailure(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "noequals"} {
		_, _, err := parseKVPair(in)
		assert.Error(t, err)
	}
}

func TestParseKVPairSuccess(t *testing.T) {
	t.Parallel()

	for expectedKey, expectedValue := range map[string]string{
		"key1": "val1",
		"key2": "=val2",
		"key3": "==val3",
		"key4": "=val4=",
		"key5": "☺",
	} {
		pair := expectedKey + "=" + expectedValue
		actualKey, actualValue, err := parseKVPair(pair)

		assert.NoError(t, err)
		assert.Equal(t, expectedKey, actualKey)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestGetParamsFromFlags(t *testing.T) {
	flagSet := setupFlags("key1=1,key2=2", `{"key3":"3","key4":"4"}`)
	values, errors := getParamsFromFlags(flagSet)
	assert.Empty(t, errors)

	// compare to expected values
	expected := render.Values{
		"key1": "1", "key2": "2", "key3": "3", "key4": "4",
	}
	assert.EqualValues(t, expected, values)
}

func TestDuplicateParameters(t *testing.T) {
	// test that duplicates in --params are detected
	flagSet := setupFlags("key1=1,key1=2", "")
	values, errors := getParamsFromFlags(flagSet)
	assert.Len(t, values, 1)
	assert.Len(t, errors, 1)

	flagSet = setupFlags("", `{"key1":"val1","key1":"2"}`)
	values, errors = getParamsFromFlags(flagSet)
	assert.Len(t, values, 1)
	assert.Len(t, errors, 0) // golang ignore duplicate JSON keys

	// test that duplicates between --params and --paramJSON are detected
	flagSet = setupFlags("key1=1,", `{"key1":"2"}`)
	values, errors = getParamsFromFlags(flagSet)
	assert.Len(t, values, 1)
	assert.Len(t, errors, 1)
}

// test that defining -p multiple times results in multiple parameters
func TestMultipleArgs(t *testing.T) {
	flagSet := pflag.NewFlagSet("", pflag.PanicOnError)
	registerParamsFlags(flagSet)
	assert.NoError(t, flagSet.Parse([]string{"-p", "key1=1", "-p", "key2=2"}))
	values, errors := getParamsFromFlags(flagSet)
	assert.Len(t, values, 2)
	assert.Len(t, errors, 0)
}
