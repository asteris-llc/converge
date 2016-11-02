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
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/shell/query"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func Test_Query_ImplementsTaskInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(query.Query))
}

func Test_Apply_ReturnsError(t *testing.T) {
	t.Parallel()
	sh := testQuery()
	_, actual := sh.Apply(context.Background())
	assert.Error(t, actual)
}

// Test Utils
func testQuery() *query.Query {
	return &query.Query{Shell: &shell.Shell{}}
}
