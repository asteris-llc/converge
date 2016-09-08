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

package param_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(param.Preparer))
}

func TestPreparerDefault(t *testing.T) {
	t.Parallel()

	prep := &param.Preparer{Default: newDefault("x")}

	result, err := prep.Prepare(fakerenderer.New())

	resultParam, ok := result.(*param.Param)
	require.True(t, ok, fmt.Sprintf("expected %T, got %T", resultParam, result))

	require.Nil(t, err)
	assert.Equal(t, *prep.Default, resultParam.Value)
}

func TestPreparerProvided(t *testing.T) {
	t.Parallel()

	prep := &param.Preparer{Default: newDefault("x")}

	result, err := prep.Prepare(fakerenderer.NewWithValue("y"))

	resultParam, ok := result.(*param.Param)
	require.True(t, ok, fmt.Sprintf("expected %T, got %T", resultParam, result))

	require.Nil(t, err)
	assert.Equal(t, "y", resultParam.Value)
}

func TestPreparerRequired(t *testing.T) {
	t.Parallel()

	prep := new(param.Preparer)
	_, err := prep.Prepare(fakerenderer.New())

	if assert.Error(t, err) {
		assert.EqualError(t, err, "param is required")
	}
}

func TestTable_PreparerValidate(t *testing.T) {
	t.Parallel()

	test_table := []struct {
		paramType string
		expected  error
		value     string
		musts     []string
	}{
		{"string", nil, "password", nil},

		{"string", nil, "password", []string{"len . | le 4"}},

		{"string", nil, "", []string{"empty"}},
		{"string", errors.New("pred#0: expected 0, got 2"), "password", []string{"empty"}},

		{"string", nil, "password", []string{`oneOf "password"`}},
		{"string", errors.New("pred#0: expected 0, got 2"), "password", []string{`oneOf "correthorsebatterystaple"`}},

		{"string", nil, "correcthorsebatterystaple", []string{`notOneOf "password" "hunter2"`}},
		{"string", errors.New("pred#0: expected 0, got 2"), "password", []string{`notOneOf "password" "hunter2"`}},

		{"int", nil, "12", nil},
		{"int", errors.New("paramType is \"int\", but converting \"twelve\" failed"), "twelve", nil},

		{"int", nil, "12", []string{"min 3"}},
		{"int", errors.New("pred#0: expected 0, got 2"), "12", []string{"max 3"}},

		{"", nil, "hello", nil},
		{"", nil, "123", nil},
	}

	for index, test := range test_table {
		prep := &param.Preparer{Type: test.paramType, Must: test.musts}

		_, actual := prep.Prepare(fakerenderer.NewWithValue(test.value))
		assert.Equal(t, test.expected, actual, fmt.Sprintf("Test #%d failed\n", index))
	}
}

func newDefault(x string) *string {
	return &x
}
