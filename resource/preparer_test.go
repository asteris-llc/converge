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

package resource_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestPreparerPrepare tests the Unmarshalling of Preparer into Resource
// structs.
func TestPreparerPrepare(t *testing.T) {
	defer logging.HideLogs(t)()

	// newWithField is a little utility. We have this pattern over and over
	// where we need to set a target field and then check that we render without
	// errors. This just encapsulates that logic (and error handling) so we
	// don't have to repeat it so much.
	newWithField := func(t *testing.T, key string, value interface{}) *testPreparerTarget {
		target := new(testPreparerTarget)
		prep := &resource.Preparer{
			Source: map[string]interface{}{
				key: value,
			},
			Destination: target,
		}

		_, err := prep.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err, "%v", err)

		return target
	}

	// strings are table stakes, so let's start with those...
	t.Run("string", func(t *testing.T) {
		target := newWithField(t, "string", "a")
		assert.Equal(t, "a", target.String)
	})

	// lists of strings are also important (argument lists, etc)
	t.Run("strings", func(t *testing.T) {
		target := newWithField(t, "strings", []string{"a"})
		assert.Equal(t, []string{"a"}, target.Strings)
	})

	// We're only testing maps with strings and bools in the tested slot, but
	// this should work with any value.
	t.Run("maps", func(t *testing.T) {
		t.Run("string-key", func(t *testing.T) {
			value := map[string]interface{}{"x": 1}
			target := newWithField(t, "stringmapkey", value)
			assert.Equal(t, value, target.StringMapKey)
		})

		t.Run("bool-key", func(t *testing.T) {
			value := map[bool]interface{}{true: 1}
			target := newWithField(t, "boolmapkey", value)
			assert.Equal(t, value, target.BoolMapKey)
		})

		t.Run("string-value", func(t *testing.T) {
			value := map[interface{}]string{1: "x"}
			target := newWithField(t, "stringmapvalue", value)
			assert.Equal(t, value, target.StringMapValue)
		})

		t.Run("bool-value", func(t *testing.T) {
			value := map[interface{}]bool{1: true}
			target := newWithField(t, "boolmapvalue", value)
			assert.Equal(t, value, target.BoolMapValue)
		})
	})

	// test time.Duration with both int64 and string
	t.Run("duration", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			target := newWithField(t, "duration", nil)
			assert.Equal(t, time.Duration(0), target.Duration)
		})

		t.Run("int64", func(t *testing.T) {
			target := newWithField(t, "duration", 1)
			assert.Equal(t, time.Duration(1), target.Duration)
		})

		t.Run("string", func(t *testing.T) {
			duration, err := time.ParseDuration("1h")
			require.NoError(t, err)
			target := newWithField(t, "duration", "1h")
			assert.Equal(t, duration, target.Duration)
		})

		t.Run("invalid", func(t *testing.T) {
			t.Run("string", func(t *testing.T) {
				val := "1"
				prep := &resource.Preparer{
					Source:      map[string]interface{}{"duration": val},
					Destination: new(testPreparerTarget),
				}

				_, err := prep.Prepare(context.Background(), fakerenderer.New())
				assert.EqualError(t, err, fmt.Sprintf("could not convert %s to duration: time: missing unit in duration %s", val, val))
			})

			t.Run("unknown", func(t *testing.T) {
				val := 3.2
				prep := &resource.Preparer{
					Source:      map[string]interface{}{"duration": val},
					Destination: new(testPreparerTarget),
				}

				_, err := prep.Prepare(context.Background(), fakerenderer.New())
				assert.EqualError(t, err, fmt.Sprintf("cannot handle duration conversion of %v", reflect.ValueOf(val).Kind()))
			})
		})
	})

	// boolean values are special. We want to support a bunch of different cases
	// and forms of truth values, so we're going to test them all in a table
	// here.
	t.Run("bool", func(t *testing.T) {
		truthTable := []struct {
			val   interface{}
			truth bool
		}{
			// true - any casing of "true" or "t", or the boolean value
			{true, true},
			{"true", true},
			{"TRUE", true},
			{"t", true},
			{"T", true},

			// false
			{false, false},
			{"false", false},
			{"FALSE", false},
			{"f", false},
			{"F", false},
			{"bananas", false}, // or anything other string except as defined above
		}

		for _, pair := range truthTable {
			t.Run(fmt.Sprintf("%T-%v", pair.val, pair.val), func(t *testing.T) {
				target := newWithField(t, "bool", pair.val)
				assert.Equal(t, pair.truth, target.Bool)
			})

			t.Run(fmt.Sprintf("slice-of-%T-%v", pair.val, pair.val), func(t *testing.T) {
				target := newWithField(t, "bools", []interface{}{pair.val})
				assert.Equal(t, []bool{pair.truth}, target.Bools)
			})
		}
	})

	// next is our "anything" escape hatch of interface{}.
	t.Run("interface", func(t *testing.T) {
		target := newWithField(t, "anything", 1)
		assert.Equal(t, 1, target.Anything)
	})

	// numbers! We want to be able to parse either a numeric value or a string.
	numberTests := []struct {
		key   string
		value interface{}
		test  func(*testing.T, *testPreparerTarget)
	}{
		// ints
		{"int", 1, func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, 1, tpt.Int) }},
		{"int", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, 1, tpt.Int) }},

		{"int8", int8(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int8(1), tpt.Int8) }},
		{"int8", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int8(1), tpt.Int8) }},

		{"int16", int16(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int16(1), tpt.Int16) }},
		{"int16", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int16(1), tpt.Int16) }},

		{"int32", int32(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int32(1), tpt.Int32) }},
		{"int32", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int32(1), tpt.Int32) }},

		{"int64", int64(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int64(1), tpt.Int64) }},
		{"int64", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, int64(1), tpt.Int64) }},

		// uints
		{"uint", 1, func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint(1), tpt.Uint) }},
		{"uint", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint(1), tpt.Uint) }},

		{"uint8", uint8(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint8(1), tpt.Uint8) }},
		{"uint8", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint8(1), tpt.Uint8) }},

		{"uint16", uint16(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint16(1), tpt.Uint16) }},
		{"uint16", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint16(1), tpt.Uint16) }},

		{"uint32", uint32(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint32(1), tpt.Uint32) }},
		{"uint32", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint32(1), tpt.Uint32) }},

		{"uint64", uint64(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint64(1), tpt.Uint64) }},
		{"uint64", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, uint64(1), tpt.Uint64) }},

		// floats
		{"float32", float32(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, float32(1), tpt.Float32) }},
		{"float32", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, float32(1), tpt.Float32) }},

		{"float64", float64(1), func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, float64(1), tpt.Float64) }},
		{"float64", "1", func(t *testing.T, tpt *testPreparerTarget) { assert.Equal(t, float64(1), tpt.Float64) }},
	}
	for _, test := range numberTests {
		t.Run(fmt.Sprintf("%s-%T", test.key, test.value), func(t *testing.T) {
			test.test(t, newWithField(t, test.key, test.value))
		})
	}

	// pointers
	t.Run("pointers", func(t *testing.T) {
		val := "test"
		target := newWithField(t, "pointer", val)
		if assert.NotNil(t, target.Pointer) {
			assert.Equal(t, *target.Pointer, val)
		}
	})

	// we do some very basic validations, let's test those too
	t.Run("valid_values", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			target := newWithField(t, "valid_values", "a")
			assert.Equal(t, "a", target.ValidValues)
		})

		t.Run("invalid", func(t *testing.T) {
			prep := &resource.Preparer{
				Source:      map[string]interface{}{"valid_values": "invalid"},
				Destination: new(testPreparerTarget),
			}

			_, err := prep.Prepare(context.Background(), fakerenderer.New())
			assert.EqualError(t, err, "value did not pass validation. Must be one of \"a\", was \"invalid\"")
		})
	})

	// type aliases are important for enum-like behavior
	t.Run("alias", func(t *testing.T) {
		target := newWithField(t, "alias", "a")
		assert.Equal(t, testAlias("a"), target.Alias)
	})

	// parameters can be required
	t.Run("required", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			prep := &resource.Preparer{
				Source:      map[string]interface{}{"required": "a"},
				Destination: new(testRequiredTarget),
			}

			_, err := prep.Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})

		t.Run("invalid", func(t *testing.T) {
			prep := &resource.Preparer{
				Source:      map[string]interface{}{},
				Destination: new(testRequiredTarget),
			}

			_, err := prep.Prepare(context.Background(), fakerenderer.New())
			assert.EqualError(t, err, `"required" is required`)
		})
	})

	// two parameters can also be mutually exclusive
	t.Run("mutually_exclusive", func(t *testing.T) {
		t.Run("invalid", func(t *testing.T) {
			prep := &resource.Preparer{
				Source: map[string]interface{}{
					"a": 1,
					"b": 2,
				},
				Destination: new(testMutuallyExclusiveTarget),
			}

			_, err := prep.Prepare(context.Background(), fakerenderer.New())
			assert.EqualError(t, err, `only one of "a" or "b" can be set`)
		})
	})
}

// testAlias is a type alias... can we deserialize those?
type testAlias string

// testPreparerTarget is a big 'ol bucket for tested values. See comments in
// TestPreparerPrepare for how these are being used.
type testPreparerTarget struct {
	String         string                 `hcl:"string"`
	Strings        []string               `hcl:"strings"`
	StringMapKey   map[string]interface{} `hcl:"stringmapkey"`
	StringMapValue map[interface{}]string `hcl:"stringmapvalue"`

	Duration time.Duration `hcl:"duration"`

	Bool         bool                 `hcl:"bool"`
	Bools        []bool               `hcl:"bools"`
	BoolMapKey   map[bool]interface{} `hcl:"boolmapkey"`
	BoolMapValue map[interface{}]bool `hcl:"boolmapvalue"`

	Anything interface{} `hcl:"anything"`

	// one of each numeric type
	Int     int     `hcl:"int"`
	Int8    int8    `hcl:"int8"`
	Int16   int16   `hcl:"int16"`
	Int32   int32   `hcl:"int32"`
	Int64   int64   `hcl:"int64"`
	Uint    uint    `hcl:"uint"`
	Uint8   uint8   `hcl:"uint8"`
	Uint16  uint16  `hcl:"uint16"`
	Uint32  uint32  `hcl:"uint32"`
	Uint64  uint64  `hcl:"uint64"`
	Float32 float32 `hcl:"float32"`
	Float64 float64 `hcl:"float64"`

	// simple validation
	ValidValues string `hcl:"valid_values" valid_values:"a"`

	// aliasing
	Alias testAlias `hcl:"alias"`

	// pointers
	Pointer *string `hcl:"pointer"`
}

func (tpt *testPreparerTarget) Prepare(context.Context, resource.Renderer) (resource.Task, error) {
	return tpt, nil
}
func (tpt *testPreparerTarget) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return nil, nil
}
func (tpt *testPreparerTarget) Apply(context.Context) (resource.TaskStatus, error) { return nil, nil }

// testRequiredTarget tests required fields. Those are invalid when empty, so
// we've got to include it separately
type testRequiredTarget struct {
	Required string `hcl:"required" required:"true"`
}

func (tpt *testRequiredTarget) Prepare(context.Context, resource.Renderer) (resource.Task, error) {
	return tpt, nil
}
func (tpt *testRequiredTarget) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return nil, nil
}
func (tpt *testRequiredTarget) Apply(context.Context) (resource.TaskStatus, error) { return nil, nil }

// testMutuallyExclusiveTarget tests mutually_exclusive fields. Those are
// invalid when empty, so we've got to include it separately
type testMutuallyExclusiveTarget struct {
	A string `hcl:"a" mutually_exclusive:"a,b"`
	B string `hcl:"b" mutually_exclusive:"a,b"`
}

func (tpt *testMutuallyExclusiveTarget) Prepare(context.Context, resource.Renderer) (resource.Task, error) {
	return tpt, nil
}
func (tpt *testMutuallyExclusiveTarget) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return nil, nil
}
func (tpt *testMutuallyExclusiveTarget) Apply(context.Context) (resource.TaskStatus, error) {
	return nil, nil
}
