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

	"github.com/asteris-llc/converge/resource"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

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
	viper.Reset()

	viper.Set("params", []string{"key1=1", "key2=2", "key3=3"})
	viper.Set("paramsJson", `{"key4": "4", "key5": "5"}`)

	values, errors := getParamsFromFlags()
	assert.Empty(t, errors)

	expected := resource.Values(map[string]resource.Value{
		"key1": "1", "key2": "2", "key3": "3", "key4": "4", "key5": "5",
	})
	assert.EqualValues(t, expected, values)
}

func TestDuplicateParameters(t *testing.T) {
	viper.Reset()

	viper.Set("params", []string{"key1=1", "key1=2"})
	values, errors := getParamsFromFlags()
	assert.Len(t, values, 1)
	assert.Len(t, errors, 1)

	viper.Set("jsonParams", `{"key1": 1, "key1", "2"}`)
	values, errors = getParamsFromFlags()
	assert.Len(t, values, 1)
	assert.Len(t, errors, 1)
}
