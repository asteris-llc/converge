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

package query_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/query"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var any = mock.Anything

func Test_Query_ImplementsTaskInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(query.Query))
}

func Test_Check_WhenRunReturnsError_ReturnsError(t *testing.T) {
	t.Parallel()
	expected := errors.New("test error")
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, expected)
	sh := testQuery(m)
	_, actual := sh.Check(fakerenderer.New())
	assert.Error(t, actual)
}

func Test_Apply_ReturnsError(t *testing.T) {
	t.Parallel()
	m := new(MockExecutor)
	sh := testQuery(m)
	_, actual := sh.Apply(fakerenderer.New())
	assert.Error(t, actual)
}

// Test Utils

func testQuery(c shell.CommandExecutor) *query.Query {
	return &query.Query{CmdGenerator: c}
}

func defaultTestQuery() *query.Query {
	return testQuery(defaultExecutor())
}

type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Run(script string) (*shell.CommandResults, error) {
	args := m.Called(script)
	return args.Get(0).(*shell.CommandResults), args.Error(1)
}

func defaultExecutor() *MockExecutor {
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, nil)
	return m
}

func resultExecutor(r *shell.CommandResults) *MockExecutor {
	m := new(MockExecutor)
	m.On("Run", any).Return(r, nil)
	return m
}
