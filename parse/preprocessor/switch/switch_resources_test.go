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

package control_test

import (
	"testing"

	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/stretchr/testify/assert"
)

// TestAppendCase ensures that cases are properly appended to switches
func TestAppendCase(t *testing.T) {
	s := &control.SwitchTask{}
	assert.Equal(t, 0, len(s.Cases()))
	s.AppendCase(&control.CaseTask{})
	assert.Equal(t, 1, len(s.Cases()))
}

// TestSortCases ensures that cases are sorted by Branch order
func TestSortCases(t *testing.T) {
	case1 := &control.CaseTask{Name: "a"}
	case2 := &control.CaseTask{Name: "b"}
	case3 := &control.CaseTask{Name: "c"}
	s := &control.SwitchTask{Branches: []string{"a", "b", "c"}}
	s.AppendCase(case2)
	s.AppendCase(case3)
	s.AppendCase(case1)
	assert.Equal(t, []*control.CaseTask{case2, case3, case1}, s.Cases())
	s.SortCases()
	assert.Equal(t, []*control.CaseTask{case1, case2, case3}, s.Cases())
}
