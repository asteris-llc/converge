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
		paramType param.ParamType
		assertion func(t assert.TestingT, object interface{}, msgAndArgs ...interface{}) bool
		value     string
		musts     []string
	}{
		// ParamTypeString checks, with pass/fail pairs

		// type check only
		{param.ParamTypeString, assert.Nil, "password", nil},

		// length func
		{param.ParamTypeString, assert.Nil, "password", []string{"le 4 length"}},
		{param.ParamTypeString, assert.NotNil, "password", []string{"ge 4 length"}},

		// empty func
		{param.ParamTypeString, assert.Nil, "", []string{"empty"}},
		{param.ParamTypeString, assert.NotNil, "password", []string{"empty"}},

		// oneOf func
		{param.ParamTypeString, assert.Nil, "password", []string{"oneOf `password`"}},
		{param.ParamTypeString, assert.NotNil, "password", []string{"oneOf `correthorsebatterystaple`"}},

		// notOneOf func
		{param.ParamTypeString, assert.Nil, "correcthorsebatterystaple", []string{"notOneOf `password hunter2`"}},
		{param.ParamTypeString, assert.NotNil, "password", []string{"notOneOf `password hunter2`"}},

		// ParamTypeInt checks, with pass/fail pairs

		// type checking only
		{param.ParamTypeInt, assert.Nil, "12", nil},
		{param.ParamTypeInt, assert.NotNil, "twelve", nil},

		// min func
		{param.ParamTypeInt, assert.Nil, "12", []string{"min 3"}},
		{param.ParamTypeInt, assert.NotNil, "12", []string{"min 48"}},

		// max func
		{param.ParamTypeInt, assert.Nil, "12", []string{"max 48"}},
		{param.ParamTypeInt, assert.NotNil, "12", []string{"max 3"}},

		// ParamTypeInferred checks

		{param.ParamTypeInferred, assert.Nil, "hello", nil},
		{param.ParamTypeInferred, assert.Nil, "123", nil},
	}

	for index, test := range test_table {
		failureMsg := fmt.Sprintf("Test #%d failed\n", index)
		prep := &param.Preparer{Type: test.paramType, Must: test.musts}
		_, actual := prep.Prepare(fakerenderer.NewWithValue(test.value))
		test.assertion(t, actual, failureMsg)
	}
}

func TestParamValidationErrorMessages(t *testing.T) {
	// Asserts that parameter validation error messages match up with the specification
}

func newDefault(x string) *string {
	return &x
}
