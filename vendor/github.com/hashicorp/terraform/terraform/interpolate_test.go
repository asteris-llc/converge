package terraform

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/config"
)

func TestInterpolater_countIndex(t *testing.T) {
	i := &Interpolater{}

	scope := &InterpolationScope{
		Path:     rootModulePath,
		Resource: &Resource{CountIndex: 42},
	}

	testInterpolate(t, i, scope, "count.index", ast.Variable{
		Value: 42,
		Type:  ast.TypeInt,
	})
}

func TestInterpolater_countIndexInWrongContext(t *testing.T) {
	i := &Interpolater{}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	n := "count.index"

	v, err := config.NewInterpolatedVariable(n)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedErr := fmt.Errorf("foo: count.index is only valid within resources")

	_, err = i.Values(scope, map[string]config.InterpolatedVariable{
		"foo": v,
	})

	if !reflect.DeepEqual(expectedErr, err) {
		t.Fatalf("expected: %#v, got %#v", expectedErr, err)
	}
}

func TestInterpolater_moduleVariable(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
						},
					},
				},
			},
			&ModuleState{
				Path: []string{RootModuleName, "child"},
				Outputs: map[string]*OutputState{
					"foo": &OutputState{
						Type:  "string",
						Value: "bar",
					},
				},
			},
		},
	}

	i := &Interpolater{
		State:     state,
		StateLock: lock,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	testInterpolate(t, i, scope, "module.child.foo", ast.Variable{
		Value: "bar",
		Type:  ast.TypeString,
	})
}

func TestInterpolater_pathCwd(t *testing.T) {
	i := &Interpolater{}
	scope := &InterpolationScope{}

	expected, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	testInterpolate(t, i, scope, "path.cwd", ast.Variable{
		Value: expected,
		Type:  ast.TypeString,
	})
}

func TestInterpolater_pathModule(t *testing.T) {
	mod := testModule(t, "interpolate-path-module")
	i := &Interpolater{
		Module: mod,
	}
	scope := &InterpolationScope{
		Path: []string{RootModuleName, "child"},
	}

	path := mod.Child([]string{"child"}).Config().Dir
	testInterpolate(t, i, scope, "path.module", ast.Variable{
		Value: path,
		Type:  ast.TypeString,
	})
}

func TestInterpolater_pathRoot(t *testing.T) {
	mod := testModule(t, "interpolate-path-module")
	i := &Interpolater{
		Module: mod,
	}
	scope := &InterpolationScope{
		Path: []string{RootModuleName, "child"},
	}

	path := mod.Config().Dir
	testInterpolate(t, i, scope, "path.root", ast.Variable{
		Value: path,
		Type:  ast.TypeString,
	})
}

func TestInterpolater_resourceVariableMap(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"amap.%":    "3",
								"amap.key1": "value1",
								"amap.key2": "value2",
								"amap.key3": "value3",
							},
						},
					},
				},
			},
		},
	}

	i := &Interpolater{
		Module:    testModule(t, "interpolate-resource-variable"),
		State:     state,
		StateLock: lock,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	testInterpolate(t, i, scope, "aws_instance.web.amap",
		interfaceToVariableSwallowError(expected))
}

func TestInterpolater_resourceVariableComplexMap(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"amap.%":      "2",
								"amap.key1.#": "2",
								"amap.key1.0": "hello",
								"amap.key1.1": "world",
								"amap.key2.#": "1",
								"amap.key2.0": "foo",
							},
						},
					},
				},
			},
		},
	}

	i := &Interpolater{
		Module:    testModule(t, "interpolate-resource-variable"),
		State:     state,
		StateLock: lock,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	expected := map[string]interface{}{
		"key1": []interface{}{"hello", "world"},
		"key2": []interface{}{"foo"},
	}

	testInterpolate(t, i, scope, "aws_instance.web.amap",
		interfaceToVariableSwallowError(expected))
}

func TestInterpolater_resourceVariable(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
			},
		},
	}

	i := &Interpolater{
		Module:    testModule(t, "interpolate-resource-variable"),
		State:     state,
		StateLock: lock,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	testInterpolate(t, i, scope, "aws_instance.web.foo", ast.Variable{
		Value: "bar",
		Type:  ast.TypeString,
	})
}

func TestInterpolater_resourceVariableMissingDuringInput(t *testing.T) {
	// During the input walk, computed resource attributes may be entirely
	// absent since we've not yet produced diffs that tell us what computed
	// attributes to expect. In that case, interpolator tolerates it and
	// indicates the value is computed.

	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path:      rootModulePath,
				Resources: map[string]*ResourceState{
				// No resources at all yet, because we're still dealing
				// with input and so the resources haven't been created.
				},
			},
		},
	}

	{
		i := &Interpolater{
			Operation: walkInput,
			Module:    testModule(t, "interpolate-resource-variable"),
			State:     state,
			StateLock: lock,
		}

		scope := &InterpolationScope{
			Path: rootModulePath,
		}

		testInterpolate(t, i, scope, "aws_instance.web.foo", ast.Variable{
			Value: config.UnknownVariableValue,
			Type:  ast.TypeString,
		})
	}

	// This doesn't apply during other walks, like plan
	{
		i := &Interpolater{
			Operation: walkPlan,
			Module:    testModule(t, "interpolate-resource-variable"),
			State:     state,
			StateLock: lock,
		}

		scope := &InterpolationScope{
			Path: rootModulePath,
		}

		testInterpolateErr(t, i, scope, "aws_instance.web.foo")
	}
}

func TestInterpolater_resourceVariableMulti(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": config.UnknownVariableValue,
							},
						},
					},
				},
			},
		},
	}

	i := &Interpolater{
		Module:    testModule(t, "interpolate-resource-variable"),
		State:     state,
		StateLock: lock,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	testInterpolate(t, i, scope, "aws_instance.web.*.foo", ast.Variable{
		Value: config.UnknownVariableValue,
		Type:  ast.TypeString,
	})
}

func interfaceToVariableSwallowError(input interface{}) ast.Variable {
	variable, _ := hil.InterfaceToVariable(input)
	return variable
}

func TestInterpolator_resourceMultiAttributes(t *testing.T) {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_route53_zone.yada": {
						Type:         "aws_route53_zone",
						Dependencies: []string{},
						Primary: &InstanceState{
							ID: "AAABBBCCCDDDEEE",
							Attributes: map[string]string{
								"name_servers.#": "4",
								"name_servers.0": "ns-1334.awsdns-38.org",
								"name_servers.1": "ns-1680.awsdns-18.co.uk",
								"name_servers.2": "ns-498.awsdns-62.com",
								"name_servers.3": "ns-601.awsdns-11.net",
								"listeners.#":    "1",
								"listeners.0":    "red",
								"tags.%":         "1",
								"tags.Name":      "reindeer",
								"nothing.#":      "0",
							},
						},
					},
				},
			},
		},
	}

	i := &Interpolater{
		Module:    testModule(t, "interpolate-multi-vars"),
		StateLock: lock,
		State:     state,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	name_servers := []interface{}{
		"ns-1334.awsdns-38.org",
		"ns-1680.awsdns-18.co.uk",
		"ns-498.awsdns-62.com",
		"ns-601.awsdns-11.net",
	}

	// More than 1 element
	testInterpolate(t, i, scope, "aws_route53_zone.yada.name_servers",
		interfaceToVariableSwallowError(name_servers))

	// Exactly 1 element
	testInterpolate(t, i, scope, "aws_route53_zone.yada.listeners",
		interfaceToVariableSwallowError([]interface{}{"red"}))

	// Zero elements
	testInterpolate(t, i, scope, "aws_route53_zone.yada.nothing",
		interfaceToVariableSwallowError([]interface{}{}))

	// Maps still need to work
	testInterpolate(t, i, scope, "aws_route53_zone.yada.tags.Name", ast.Variable{
		Value: "reindeer",
		Type:  ast.TypeString,
	})
}

func TestInterpolator_resourceMultiAttributesWithResourceCount(t *testing.T) {
	i := getInterpolaterFixture(t)
	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	name_servers := []interface{}{
		"ns-1334.awsdns-38.org",
		"ns-1680.awsdns-18.co.uk",
		"ns-498.awsdns-62.com",
		"ns-601.awsdns-11.net",
		"ns-000.awsdns-38.org",
		"ns-444.awsdns-18.co.uk",
		"ns-999.awsdns-62.com",
		"ns-666.awsdns-11.net",
	}

	// More than 1 element
	testInterpolate(t, i, scope, "aws_route53_zone.terra.0.name_servers",
		interfaceToVariableSwallowError(name_servers[:4]))

	// More than 1 element in both
	testInterpolate(t, i, scope, "aws_route53_zone.terra.*.name_servers",
		interfaceToVariableSwallowError([]interface{}{name_servers[:4], name_servers[4:]}))

	// Exactly 1 element
	testInterpolate(t, i, scope, "aws_route53_zone.terra.0.listeners",
		interfaceToVariableSwallowError([]interface{}{"red"}))

	// Exactly 1 element in both
	testInterpolate(t, i, scope, "aws_route53_zone.terra.*.listeners",
		interfaceToVariableSwallowError([]interface{}{[]interface{}{"red"}, []interface{}{"blue"}}))

	// Zero elements
	testInterpolate(t, i, scope, "aws_route53_zone.terra.0.nothing",
		interfaceToVariableSwallowError([]interface{}{}))

	// Zero + 1 element
	testInterpolate(t, i, scope, "aws_route53_zone.terra.*.special",
		interfaceToVariableSwallowError([]interface{}{[]interface{}{"extra"}}))

	// Maps still need to work
	testInterpolate(t, i, scope, "aws_route53_zone.terra.0.tags.Name", ast.Variable{
		Value: "reindeer",
		Type:  ast.TypeString,
	})

	// Maps still need to work in both
	testInterpolate(t, i, scope, "aws_route53_zone.terra.*.tags.Name",
		interfaceToVariableSwallowError([]interface{}{"reindeer", "white-hart"}))
}

func TestInterpolator_resourceMultiAttributesComputed(t *testing.T) {
	lock := new(sync.RWMutex)
	// The state would never be written with an UnknownVariableValue in it, but
	// it can/does exist that way in memory during the plan phase.
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_route53_zone.yada": &ResourceState{
						Type: "aws_route53_zone",
						Primary: &InstanceState{
							ID: "z-abc123",
							Attributes: map[string]string{
								"name_servers.#": config.UnknownVariableValue,
							},
						},
					},
				},
			},
		},
	}
	i := &Interpolater{
		Module:    testModule(t, "interpolate-multi-vars"),
		StateLock: lock,
		State:     state,
	}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	testInterpolate(t, i, scope, "aws_route53_zone.yada.name_servers", ast.Variable{
		Value: config.UnknownVariableValue,
		Type:  ast.TypeString,
	})
}

func TestInterpolater_selfVarWithoutResource(t *testing.T) {
	i := &Interpolater{}

	scope := &InterpolationScope{
		Path: rootModulePath,
	}

	v, err := config.NewInterpolatedVariable("self.name")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = i.Values(scope, map[string]config.InterpolatedVariable{"foo": v})
	if err == nil {
		t.Fatalf("expected err, got none")
	}
}

func getInterpolaterFixture(t *testing.T) *Interpolater {
	lock := new(sync.RWMutex)
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_route53_zone.terra.0": &ResourceState{
						Type:         "aws_route53_zone",
						Dependencies: []string{},
						Primary: &InstanceState{
							ID: "AAABBBCCCDDDEEE",
							Attributes: map[string]string{
								"name_servers.#": "4",
								"name_servers.0": "ns-1334.awsdns-38.org",
								"name_servers.1": "ns-1680.awsdns-18.co.uk",
								"name_servers.2": "ns-498.awsdns-62.com",
								"name_servers.3": "ns-601.awsdns-11.net",
								"listeners.#":    "1",
								"listeners.0":    "red",
								"tags.%":         "1",
								"tags.Name":      "reindeer",
								"nothing.#":      "0",
							},
						},
					},
					"aws_route53_zone.terra.1": &ResourceState{
						Type:         "aws_route53_zone",
						Dependencies: []string{},
						Primary: &InstanceState{
							ID: "EEEFFFGGGHHHIII",
							Attributes: map[string]string{
								"name_servers.#": "4",
								"name_servers.0": "ns-000.awsdns-38.org",
								"name_servers.1": "ns-444.awsdns-18.co.uk",
								"name_servers.2": "ns-999.awsdns-62.com",
								"name_servers.3": "ns-666.awsdns-11.net",
								"listeners.#":    "1",
								"listeners.0":    "blue",
								"special.#":      "1",
								"special.0":      "extra",
								"tags.%":         "1",
								"tags.Name":      "white-hart",
								"nothing.#":      "0",
							},
						},
					},
				},
			},
		},
	}

	return &Interpolater{
		Module:    testModule(t, "interpolate-multi-vars"),
		StateLock: lock,
		State:     state,
	}
}

func testInterpolate(
	t *testing.T, i *Interpolater,
	scope *InterpolationScope,
	n string, expectedVar ast.Variable) {
	v, err := config.NewInterpolatedVariable(n)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual, err := i.Values(scope, map[string]config.InterpolatedVariable{
		"foo": v,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]ast.Variable{
		"foo": expectedVar,
	}
	if !reflect.DeepEqual(actual, expected) {
		spew.Config.DisableMethods = true
		t.Fatalf("%q:\n\n  actual: %#v\nexpected: %#v\n\n%s\n\n%s\n\n", n, actual, expected,
			spew.Sdump(actual), spew.Sdump(expected))
	}
}

func testInterpolateErr(
	t *testing.T, i *Interpolater,
	scope *InterpolationScope,
	n string) {
	v, err := config.NewInterpolatedVariable(n)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = i.Values(scope, map[string]config.InterpolatedVariable{
		"foo": v,
	})
	if err == nil {
		t.Fatalf("%q: succeeded, but wanted error", n)
	}
}
