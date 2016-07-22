package terraform

import (
	"testing"
)

func TestStateAdd(t *testing.T) {
	cases := map[string]struct {
		Err      bool
		From, To string
		Value    interface{}
		One, Two *State
	}{
		"ModuleState => Module Addr (new)": {
			false,
			"",
			"module.foo",
			&ModuleState{
				Path: rootModulePath,
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

			&State{},
			&State{
				Modules: []*ModuleState{
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
		},

		"ModuleState => Nested Module Addr (new)": {
			false,
			"",
			"module.foo.module.bar",
			&ModuleState{
				Path: rootModulePath,
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

			&State{},
			&State{
				Modules: []*ModuleState{
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
		},

		"ModuleState w/ outputs and deps => Module Addr (new)": {
			false,
			"",
			"module.foo",
			&ModuleState{
				Path: rootModulePath,
				Outputs: map[string]*OutputState{
					"foo": &OutputState{
						Type:      "string",
						Sensitive: false,
						Value:     "bar",
					},
				},
				Dependencies: []string{"foo"},
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

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root", "foo"},
						Outputs: map[string]*OutputState{
							"foo": &OutputState{
								Type:      "string",
								Sensitive: false,
								Value:     "bar",
							},
						},
						Dependencies: []string{"foo"},
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
		},

		"ModuleState => Module Addr (existing)": {
			true,
			"",
			"module.foo",
			&ModuleState{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"test_instance.baz": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			nil,
		},

		"ResourceState => Resource Addr (new)": {
			false,
			"aws_instance.bar",
			"aws_instance.foo",
			&ResourceState{
				Type: "test_instance",
				Primary: &InstanceState{
					ID: "foo",
				},
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root"},
						Resources: map[string]*ResourceState{
							"aws_instance.foo": &ResourceState{
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

		"ResourceState w/ deps, provider => Resource Addr (new)": {
			false,
			"aws_instance.bar",
			"aws_instance.foo",
			&ResourceState{
				Type:         "test_instance",
				Provider:     "foo",
				Dependencies: []string{"bar"},
				Primary: &InstanceState{
					ID: "foo",
				},
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root"},
						Resources: map[string]*ResourceState{
							"aws_instance.foo": &ResourceState{
								Type:         "test_instance",
								Provider:     "foo",
								Dependencies: []string{"bar"},
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"ResourceState tainted => Resource Addr (new)": {
			false,
			"aws_instance.bar",
			"aws_instance.foo",
			&ResourceState{
				Type: "test_instance",
				Primary: &InstanceState{
					ID:      "foo",
					Tainted: true,
				},
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root"},
						Resources: map[string]*ResourceState{
							"aws_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID:      "foo",
									Tainted: true,
								},
							},
						},
					},
				},
			},
		},

		"ResourceState => Resource Addr (existing)": {
			true,
			"aws_instance.bar",
			"aws_instance.foo",
			&ResourceState{
				Type: "test_instance",
				Primary: &InstanceState{
					ID: "foo",
				},
			},

			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root"},
						Resources: map[string]*ResourceState{
							"aws_instance.foo": &ResourceState{
								Type: "test_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			nil,
		},

		"ResourceState => Module (new)": {
			false,
			"aws_instance.bar",
			"module.foo",
			&ResourceState{
				Type: "test_instance",
				Primary: &InstanceState{
					ID: "foo",
				},
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"aws_instance.bar": &ResourceState{
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

		"InstanceState => Resource (new)": {
			false,
			"aws_instance.bar.primary",
			"aws_instance.baz",
			&InstanceState{
				ID: "foo",
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root"},
						Resources: map[string]*ResourceState{
							"aws_instance.baz": &ResourceState{
								Type: "aws_instance",
								Primary: &InstanceState{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},

		"InstanceState => Module (new)": {
			false,
			"aws_instance.bar.primary",
			"module.foo",
			&InstanceState{
				ID: "foo",
			},

			&State{},
			&State{
				Modules: []*ModuleState{
					&ModuleState{
						Path: []string{"root", "foo"},
						Resources: map[string]*ResourceState{
							"aws_instance.bar": &ResourceState{
								Type: "aws_instance",
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
		// Make sure they're both initialized as normal
		tc.One.init()
		if tc.Two != nil {
			tc.Two.init()
		}

		// Add the value
		err := tc.One.Add(tc.From, tc.To, tc.Value)
		if (err != nil) != tc.Err {
			t.Fatalf("bad: %s\n\n%s", k, err)
		}
		if tc.Err {
			continue
		}

		// Prune them both to be sure
		tc.One.prune()
		tc.Two.prune()

		// Verify equality
		if !tc.One.Equal(tc.Two) {
			t.Fatalf("Bad: %s\n\n%#v\n\n%#v", k, tc.One, tc.Two)
			//t.Fatalf("Bad: %s\n\n%s\n\n%s", k, tc.One.String(), tc.Two.String())
		}
	}
}
