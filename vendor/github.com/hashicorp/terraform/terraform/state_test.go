package terraform

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/config"
)

func TestStateAddModule(t *testing.T) {
	cases := []struct {
		In  [][]string
		Out [][]string
	}{
		{
			[][]string{
				[]string{"root"},
				[]string{"root", "child"},
			},
			[][]string{
				[]string{"root"},
				[]string{"root", "child"},
			},
		},

		{
			[][]string{
				[]string{"root", "foo", "bar"},
				[]string{"root", "foo"},
				[]string{"root"},
				[]string{"root", "bar"},
			},
			[][]string{
				[]string{"root"},
				[]string{"root", "bar"},
				[]string{"root", "foo"},
				[]string{"root", "foo", "bar"},
			},
		},
		// Same last element, different middle element
		{
			[][]string{
				[]string{"root", "foo", "bar"}, // This one should sort after...
				[]string{"root", "foo"},
				[]string{"root"},
				[]string{"root", "bar", "bar"}, // ...this one.
				[]string{"root", "bar"},
			},
			[][]string{
				[]string{"root"},
				[]string{"root", "bar"},
				[]string{"root", "foo"},
				[]string{"root", "bar", "bar"},
				[]string{"root", "foo", "bar"},
			},
		},
	}

	for _, tc := range cases {
		s := new(State)
		for _, p := range tc.In {
			s.AddModule(p)
		}

		actual := make([][]string, 0, len(tc.In))
		for _, m := range s.Modules {
			actual = append(actual, m.Path)
		}

		if !reflect.DeepEqual(actual, tc.Out) {
			t.Fatalf("In: %#v\n\nOut: %#v", tc.In, actual)
		}
	}
}

func TestStateOutputTypeRoundTrip(t *testing.T) {
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: RootModulePath,
				Outputs: map[string]*OutputState{
					"string_output": &OutputState{
						Value: "String Value",
						Type:  "string",
					},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	if err := WriteState(state, buf); err != nil {
		t.Fatalf("err: %s", err)
	}

	roundTripped, err := ReadState(buf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(state, roundTripped) {
		t.Fatalf("bad: %#v", roundTripped)
	}
}

func TestStateModuleOrphans(t *testing.T) {
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: RootModulePath,
			},
			&ModuleState{
				Path: []string{RootModuleName, "foo"},
			},
			&ModuleState{
				Path: []string{RootModuleName, "bar"},
			},
		},
	}

	config := testModule(t, "state-module-orphans").Config()
	actual := state.ModuleOrphans(RootModulePath, config)
	expected := [][]string{
		[]string{RootModuleName, "foo"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestStateModuleOrphans_nested(t *testing.T) {
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: RootModulePath,
			},
			&ModuleState{
				Path: []string{RootModuleName, "foo", "bar"},
			},
		},
	}

	actual := state.ModuleOrphans(RootModulePath, nil)
	expected := [][]string{
		[]string{RootModuleName, "foo"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestStateModuleOrphans_nilConfig(t *testing.T) {
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: RootModulePath,
			},
			&ModuleState{
				Path: []string{RootModuleName, "foo"},
			},
			&ModuleState{
				Path: []string{RootModuleName, "bar"},
			},
		},
	}

	actual := state.ModuleOrphans(RootModulePath, nil)
	expected := [][]string{
		[]string{RootModuleName, "foo"},
		[]string{RootModuleName, "bar"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestStateModuleOrphans_deepNestedNilConfig(t *testing.T) {
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: RootModulePath,
			},
			&ModuleState{
				Path: []string{RootModuleName, "parent", "childfoo"},
			},
			&ModuleState{
				Path: []string{RootModuleName, "parent", "childbar"},
			},
		},
	}

	actual := state.ModuleOrphans(RootModulePath, nil)
	expected := [][]string{
		[]string{RootModuleName, "parent"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestStateDeepCopy(t *testing.T) {
	cases := []struct {
		One, Two *State
		F        func(*State) interface{}
	}{
		// Version
		{
			&State{Version: 5},
			&State{Version: 5},
			func(s *State) interface{} { return s.Version },
		},

		// TFVersion
		{
			&State{TFVersion: "5"},
			&State{TFVersion: "5"},
			func(s *State) interface{} { return s.TFVersion },
		},
	}

	for i, tc := range cases {
		actual := tc.F(tc.One.DeepCopy())
		expected := tc.F(tc.Two)
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, actual, expected)
		}
	}
}

func TestStateEqual(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *State
	}{
		// Nils
		{
			false,
			nil,
			&State{Version: 2},
		},

		{
			true,
			nil,
			nil,
		},

		// Different versions
		{
			false,
			&State{Version: 5},
			&State{Version: 2},
		},

		// Different modules
		{
			false,
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: RootModulePath,
					},
				},
			},
			&State{},
		},

		{
			true,
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: RootModulePath,
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: RootModulePath,
					},
				},
			},
		},

		// Meta differs
		{
			false,
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Primary: &InstanceState{
									Meta: map[string]string{
										"schema_version": "1",
									},
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Primary: &InstanceState{
									Meta: map[string]string{
										"schema_version": "2",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range cases {
		if tc.One.Equal(tc.Two) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
		if tc.Two.Equal(tc.One) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
	}
}

func TestStateCompareAges(t *testing.T) {
	cases := []struct {
		Result   StateAgeComparison
		Err      bool
		One, Two *State
	}{
		{
			StateAgeEqual, false,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "1",
				Serial:  2,
			},
		},
		{
			StateAgeReceiverOlder, false,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "1",
				Serial:  3,
			},
		},
		{
			StateAgeReceiverNewer, false,
			&State{
				Lineage: "1",
				Serial:  3,
			},
			&State{
				Lineage: "1",
				Serial:  2,
			},
		},
		{
			StateAgeEqual, true,
			&State{
				Lineage: "1",
				Serial:  2,
			},
			&State{
				Lineage: "2",
				Serial:  2,
			},
		},
		{
			StateAgeEqual, true,
			&State{
				Lineage: "1",
				Serial:  3,
			},
			&State{
				Lineage: "2",
				Serial:  2,
			},
		},
	}

	for i, tc := range cases {
		result, err := tc.One.CompareAges(tc.Two)

		if err != nil && !tc.Err {
			t.Errorf(
				"%d: got error, but want success\n\n%s\n\n%s",
				i, tc.One, tc.Two,
			)
			continue
		}

		if err == nil && tc.Err {
			t.Errorf(
				"%d: got success, but want error\n\n%s\n\n%s",
				i, tc.One, tc.Two,
			)
			continue
		}

		if result != tc.Result {
			t.Errorf(
				"%d: got result %d, but want %d\n\n%s\n\n%s",
				i, result, tc.Result, tc.One, tc.Two,
			)
			continue
		}
	}
}

func TestStateSameLineage(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *State
	}{
		{
			true,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "1",
			},
		},
		{
			// Empty lineage is compatible with all
			true,
			&State{
				Lineage: "",
			},
			&State{
				Lineage: "1",
			},
		},
		{
			// Empty lineage is compatible with all
			true,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "",
			},
		},
		{
			false,
			&State{
				Lineage: "1",
			},
			&State{
				Lineage: "2",
			},
		},
	}

	for i, tc := range cases {
		result := tc.One.SameLineage(tc.Two)

		if result != tc.Result {
			t.Errorf(
				"%d: got %v, but want %v\n\n%s\n\n%s",
				i, result, tc.Result, tc.One, tc.Two,
			)
			continue
		}
	}
}

func TestStateIncrementSerialMaybe(t *testing.T) {
	cases := map[string]struct {
		S1, S2 *State
		Serial int64
	}{
		"S2 is nil": {
			&State{},
			nil,
			0,
		},
		"S2 is identical": {
			&State{},
			&State{},
			0,
		},
		"S2 is different": {
			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{Path: rootModulePath},
				},
			},
			1,
		},
		"S2 is different, but only via Instance Metadata": {
			&State{
				Serial: 3,
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Primary: &InstanceState{
									Meta: map[string]string{},
								},
							},
						},
					},
				},
			},
			&State{
				Serial: 3,
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Primary: &InstanceState{
									Meta: map[string]string{
										"schema_version": "1",
									},
								},
							},
						},
					},
				},
			},
			4,
		},
		"S1 serial is higher": {
			&State{Serial: 5},
			&State{
				Serial: 3,
				Modules: []*ModuleState{
					&ModuleState{Path: rootModulePath},
				},
			},
			5,
		},
		"S2 has a different TFVersion": {
			&State{TFVersion: "0.1"},
			&State{TFVersion: "0.2"},
			1,
		},
	}

	for name, tc := range cases {
		tc.S1.IncrementSerialMaybe(tc.S2)
		if tc.S1.Serial != tc.Serial {
			t.Fatalf("Bad: %s\nGot: %d", name, tc.S1.Serial)
		}
	}
}

func TestStateRemove(t *testing.T) {
	cases := map[string]struct {
		Address  string
		One, Two *State
	}{
		"simple resource": {
			"test_instance.foo",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path:      rootModulePath,
						Resources: map[string]*ResourceState{},
					},
				},
			},
		},

		"single instance": {
			"test_instance.foo.primary",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path:      rootModulePath,
						Resources: map[string]*ResourceState{},
					},
				},
			},
		},

		"single instance in multi-count": {
			"test_instance.foo[0]",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.0": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.foo.1": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.1": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"single resource, multi-count": {
			"test_instance.foo",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo.0": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.foo.1": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path:      rootModulePath,
						Resources: map[string]*ResourceState{},
					},
				},
			},
		},

		"full module": {
			"module.foo",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					&ModuleState{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"module and children": {
			"module.foo",
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					&ModuleState{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},

					&ModuleState{
						Path: []string{"root", "foo", "bar"},
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},

							"test_instance.bar": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: rootModulePath,
						Resources: map[string]*ResourceState{
							"test_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},
	}

	for k, tc := range cases {
		if err := tc.One.Remove(tc.Address); err != nil {
			t.Fatalf("bad: %s\n\n%s", k, err)
		}

		if !tc.One.Equal(tc.Two) {
			t.Fatalf("Bad: %s\n\n%s\n\n%s", k, tc.One.String(), tc.Two.String())
		}
	}
}

func TestResourceStateEqual(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *ResourceState
	}{
		// Different types
		{
			false,
			&ResourceState{Type: "foo"},
			&ResourceState{Type: "bar"},
		},

		// Different dependencies
		{
			false,
			&ResourceState{Dependencies: []string{"foo"}},
			&ResourceState{Dependencies: []string{"bar"}},
		},

		{
			false,
			&ResourceState{Dependencies: []string{"foo", "bar"}},
			&ResourceState{Dependencies: []string{"foo"}},
		},

		{
			true,
			&ResourceState{Dependencies: []string{"bar", "foo"}},
			&ResourceState{Dependencies: []string{"foo", "bar"}},
		},

		// Different primaries
		{
			false,
			&ResourceState{Primary: nil},
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
		},

		{
			true,
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
			&ResourceState{Primary: &InstanceState{ID: "foo"}},
		},

		// Different tainted
		{
			false,
			&ResourceState{
				Primary: &InstanceState{
					ID: "foo",
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},

		{
			true,
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},
	}

	for i, tc := range cases {
		if tc.One.Equal(tc.Two) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
		if tc.Two.Equal(tc.One) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
	}
}

func TestResourceStateTaint(t *testing.T) {
	cases := map[string]struct {
		Input  *ResourceState
		Output *ResourceState
	}{
		"no primary": {
			&ResourceState{},
			&ResourceState{},
		},

		"primary, not tainted": {
			&ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},

		"primary, tainted": {
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			&ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
		},
	}

	for k, tc := range cases {
		tc.Input.Taint()
		if !reflect.DeepEqual(tc.Input, tc.Output) {
			t.Fatalf(
				"Failure: %s\n\nExpected: %#v\n\nGot: %#v",
				k, tc.Output, tc.Input)
		}
	}
}

func TestResourceStateUntaint(t *testing.T) {
	cases := map[string]struct {
		Input          *ResourceState
		ExpectedOutput *ResourceState
	}{
		"no primary, err": {
			Input:          &ResourceState{},
			ExpectedOutput: &ResourceState{},
		},

		"primary, not tainted": {
			Input: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
			ExpectedOutput: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
		},
		"primary, tainted": {
			Input: &ResourceState{
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},
			ExpectedOutput: &ResourceState{
				Primary: &InstanceState{ID: "foo"},
			},
		},
	}

	for k, tc := range cases {
		tc.Input.Untaint()
		if !reflect.DeepEqual(tc.Input, tc.ExpectedOutput) {
			t.Fatalf(
				"Failure: %s\n\nExpected: %#v\n\nGot: %#v",
				k, tc.ExpectedOutput, tc.Input)
		}
	}
}

func TestInstanceStateEmpty(t *testing.T) {
	cases := map[string]struct {
		In     *InstanceState
		Result bool
	}{
		"nil is empty": {
			nil,
			true,
		},
		"non-nil but without ID is empty": {
			&InstanceState{},
			true,
		},
		"with ID is not empty": {
			&InstanceState{
				ID: "i-abc123",
			},
			false,
		},
	}

	for tn, tc := range cases {
		if tc.In.Empty() != tc.Result {
			t.Fatalf("%q expected %#v to be empty: %#v", tn, tc.In, tc.Result)
		}
	}
}

func TestInstanceStateEqual(t *testing.T) {
	cases := []struct {
		Result   bool
		One, Two *InstanceState
	}{
		// Nils
		{
			false,
			nil,
			&InstanceState{},
		},

		{
			false,
			&InstanceState{},
			nil,
		},

		// Different IDs
		{
			false,
			&InstanceState{ID: "foo"},
			&InstanceState{ID: "bar"},
		},

		// Different Attributes
		{
			false,
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
			&InstanceState{Attributes: map[string]string{"foo": "baz"}},
		},

		// Different Attribute keys
		{
			false,
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
			&InstanceState{Attributes: map[string]string{"bar": "baz"}},
		},

		{
			false,
			&InstanceState{Attributes: map[string]string{"bar": "baz"}},
			&InstanceState{Attributes: map[string]string{"foo": "bar"}},
		},
	}

	for i, tc := range cases {
		if tc.One.Equal(tc.Two) != tc.Result {
			t.Fatalf("Bad: %d\n\n%s\n\n%s", i, tc.One.String(), tc.Two.String())
		}
	}
}

func TestStateEmpty(t *testing.T) {
	cases := []struct {
		In     *State
		Result bool
	}{
		{
			nil,
			true,
		},
		{
			&State{},
			true,
		},
		{
			&State{
				Remote: &RemoteState{Type: "foo"},
			},
			true,
		},
		{
			&State{
				Modules: []*ModuleState{
					&ModuleState{},
				},
			},
			false,
		},
	}

	for i, tc := range cases {
		if tc.In.Empty() != tc.Result {
			t.Fatalf("bad %d %#v:\n\n%#v", i, tc.Result, tc.In)
		}
	}
}

func TestStateFromFutureTerraform(t *testing.T) {
	cases := []struct {
		In     string
		Result bool
	}{
		{
			"",
			false,
		},
		{
			"0.1",
			false,
		},
		{
			"999.15.1",
			true,
		},
	}

	for _, tc := range cases {
		state := &State{TFVersion: tc.In}
		actual := state.FromFutureTerraform()
		if actual != tc.Result {
			t.Fatalf("%s: bad: %v", tc.In, actual)
		}
	}
}

func TestStateIsRemote(t *testing.T) {
	cases := []struct {
		In     *State
		Result bool
	}{
		{
			nil,
			false,
		},
		{
			&State{},
			false,
		},
		{
			&State{
				Remote: &RemoteState{Type: "foo"},
			},
			true,
		},
	}

	for i, tc := range cases {
		if tc.In.IsRemote() != tc.Result {
			t.Fatalf("bad %d %#v:\n\n%#v", i, tc.Result, tc.In)
		}
	}
}

func TestInstanceState_MergeDiff(t *testing.T) {
	is := InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"foo":  "bar",
			"port": "8000",
		},
	}

	diff := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				Old: "bar",
				New: "baz",
			},
			"bar": &ResourceAttrDiff{
				Old: "",
				New: "foo",
			},
			"baz": &ResourceAttrDiff{
				Old:         "",
				New:         "foo",
				NewComputed: true,
			},
			"port": &ResourceAttrDiff{
				NewRemoved: true,
			},
		},
	}

	is2 := is.MergeDiff(diff)

	expected := map[string]string{
		"foo": "baz",
		"bar": "foo",
		"baz": config.UnknownVariableValue,
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

func TestInstanceState_MergeDiff_nil(t *testing.T) {
	var is *InstanceState

	diff := &InstanceDiff{
		Attributes: map[string]*ResourceAttrDiff{
			"foo": &ResourceAttrDiff{
				Old: "",
				New: "baz",
			},
		},
	}

	is2 := is.MergeDiff(diff)

	expected := map[string]string{
		"foo": "baz",
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

func TestInstanceState_MergeDiff_nilDiff(t *testing.T) {
	is := InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"foo": "bar",
		},
	}

	is2 := is.MergeDiff(nil)

	expected := map[string]string{
		"foo": "bar",
	}

	if !reflect.DeepEqual(expected, is2.Attributes) {
		t.Fatalf("bad: %#v", is2.Attributes)
	}
}

func TestReadWriteState(t *testing.T) {
	state := &State{
		Serial: 9,
		Remote: &RemoteState{
			Type: "http",
			Config: map[string]string{
				"url": "http://my-cool-server.com/",
			},
		},
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Dependencies: []string{
					"aws_instance.bar",
				},
				Resources: map[string]*ResourceState{
					"foo": &ResourceState{
						Primary: &InstanceState{
							ID: "bar",
							Ephemeral: EphemeralState{
								ConnInfo: map[string]string{
									"type":     "ssh",
									"user":     "root",
									"password": "supersecret",
								},
							},
						},
					},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	if err := WriteState(state, buf); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify that the version and serial are set
	if state.Version != StateVersion {
		t.Fatalf("bad version number: %d", state.Version)
	}

	actual, err := ReadState(buf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// ReadState should not restore sensitive information!
	mod := state.RootModule()
	mod.Resources["foo"].Primary.Ephemeral = EphemeralState{}

	if !reflect.DeepEqual(actual, state) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestReadStateNewVersion(t *testing.T) {
	type out struct {
		Version int
	}

	buf, err := json.Marshal(&out{StateVersion + 1})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	s, err := ReadState(bytes.NewReader(buf))
	if s != nil {
		t.Fatalf("unexpected: %#v", s)
	}
	if !strings.Contains(err.Error(), "does not support state version") {
		t.Fatalf("err: %v", err)
	}
}

func TestReadStateTFVersion(t *testing.T) {
	type tfVersion struct {
		Version   int    `json:"version"`
		TFVersion string `json:"terraform_version"`
	}

	cases := []struct {
		Written string
		Read    string
		Err     bool
	}{
		{
			"0.0.0",
			"0.0.0",
			false,
		},
		{
			"",
			"",
			false,
		},
		{
			"bad",
			"",
			true,
		},
	}

	for _, tc := range cases {
		buf, err := json.Marshal(&tfVersion{
			Version:   2,
			TFVersion: tc.Written,
		})
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		s, err := ReadState(bytes.NewReader(buf))
		if (err != nil) != tc.Err {
			t.Fatalf("%s: err: %s", tc.Written, err)
		}
		if err != nil {
			continue
		}

		if s.TFVersion != tc.Read {
			t.Fatalf("%s: bad: %s", tc.Written, s.TFVersion)
		}
	}
}

func TestWriteStateTFVersion(t *testing.T) {
	cases := []struct {
		Write string
		Read  string
		Err   bool
	}{
		{
			"0.0.0",
			"0.0.0",
			false,
		},
		{
			"",
			"",
			false,
		},
		{
			"bad",
			"",
			true,
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		err := WriteState(&State{TFVersion: tc.Write}, &buf)
		if (err != nil) != tc.Err {
			t.Fatalf("%s: err: %s", tc.Write, err)
		}
		if err != nil {
			continue
		}

		s, err := ReadState(&buf)
		if err != nil {
			t.Fatalf("%s: err: %s", tc.Write, err)
		}

		if s.TFVersion != tc.Read {
			t.Fatalf("%s: bad: %s", tc.Write, s.TFVersion)
		}
	}
}

func TestParseResourceStateKey(t *testing.T) {
	cases := []struct {
		Input       string
		Expected    *ResourceStateKey
		ExpectedErr bool
	}{
		{
			Input: "aws_instance.foo.3",
			Expected: &ResourceStateKey{
				Mode:  config.ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: 3,
			},
		},
		{
			Input: "aws_instance.foo.0",
			Expected: &ResourceStateKey{
				Mode:  config.ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: 0,
			},
		},
		{
			Input: "aws_instance.foo",
			Expected: &ResourceStateKey{
				Mode:  config.ManagedResourceMode,
				Type:  "aws_instance",
				Name:  "foo",
				Index: -1,
			},
		},
		{
			Input: "data.aws_ami.foo",
			Expected: &ResourceStateKey{
				Mode:  config.DataResourceMode,
				Type:  "aws_ami",
				Name:  "foo",
				Index: -1,
			},
		},
		{
			Input:       "aws_instance.foo.malformed",
			ExpectedErr: true,
		},
		{
			Input:       "aws_instance.foo.malformedwithnumber.123",
			ExpectedErr: true,
		},
		{
			Input:       "malformed",
			ExpectedErr: true,
		},
	}
	for _, tc := range cases {
		rsk, err := ParseResourceStateKey(tc.Input)
		if rsk != nil && tc.Expected != nil && !rsk.Equal(tc.Expected) {
			t.Fatalf("%s: expected %s, got %s", tc.Input, tc.Expected, rsk)
		}
		if (err != nil) != tc.ExpectedErr {
			t.Fatalf("%s: expected err: %t, got %s", tc.Input, tc.ExpectedErr, err)
		}
	}
}
