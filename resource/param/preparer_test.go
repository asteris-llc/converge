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

func TestPreparerValidate(t *testing.T) {
	t.Parallel()

	test_table := []struct {
		pType string
		exp   error
		val   string
		musts []string
	}{
		// STRING TESTS

		// type check only
		{"string", nil, "password", nil},

		// rule checks
		{"string", nil, "password", []string{"len . | le 4"}},
		{"string", errors.New("Expected true from rule-0-must-0, got false"), "password", []string{"empty"}},

		// INT TESTS

		// type check only
		{"int", nil, "12", nil},
		{"int", errors.New("vtype is \"int\", but converting \"twelve\" failed"), "twelve", nil},

		// rule checks
		{"int", nil, "12", []string{"min 3"}},
		{"int", errors.New("Expected true from rule-0-must-0, got false"), "12", []string{"max 3"}},
	}

	for _, test := range test_table {
		rules := &param.Rule{Must: test.musts}
		prep := &param.Preparer{Type: test.pType, Rules: []*param.Rule{rules}}

		_, act := prep.Prepare(fakerenderer.NewWithValue(test.val))
		assert.Equal(t, test.exp, act)
	}
}

func newDefault(x string) *string {
	return &x
}
