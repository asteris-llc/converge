package schema

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
)

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = new(Provider)
}

func TestProviderConfigure(t *testing.T) {
	cases := []struct {
		P      *Provider
		Config map[string]interface{}
		Err    bool
	}{
		{
			P:      &Provider{},
			Config: nil,
			Err:    false,
		},

		{
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": &Schema{
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureFunc: func(d *ResourceData) (interface{}, error) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, fmt.Errorf("nope")
				},
			},
			Config: map[string]interface{}{
				"foo": 42,
			},
			Err: false,
		},

		{
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": &Schema{
						Type:     TypeInt,
						Optional: true,
					},
				},

				ConfigureFunc: func(d *ResourceData) (interface{}, error) {
					if d.Get("foo").(int) == 42 {
						return nil, nil
					}

					return nil, fmt.Errorf("nope")
				},
			},
			Config: map[string]interface{}{
				"foo": 52,
			},
			Err: true,
		},
	}

	for i, tc := range cases {
		c, err := config.NewRawConfig(tc.Config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		err = tc.P.Configure(terraform.NewResourceConfig(c))
		if err != nil != tc.Err {
			t.Fatalf("%d: %s", i, err)
		}
	}
}

func TestProviderResources(t *testing.T) {
	cases := []struct {
		P      *Provider
		Result []terraform.ResourceType
	}{
		{
			P:      &Provider{},
			Result: []terraform.ResourceType{},
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": nil,
				},
			},
			Result: []terraform.ResourceType{
				terraform.ResourceType{Name: "bar"},
				terraform.ResourceType{Name: "foo"},
			},
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": &Resource{Importer: &ResourceImporter{}},
					"baz": nil,
				},
			},
			Result: []terraform.ResourceType{
				terraform.ResourceType{Name: "bar", Importable: true},
				terraform.ResourceType{Name: "baz"},
				terraform.ResourceType{Name: "foo"},
			},
		},
	}

	for i, tc := range cases {
		actual := tc.P.Resources()
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d: %#v", i, actual)
		}
	}
}

func TestProviderDataSources(t *testing.T) {
	cases := []struct {
		P      *Provider
		Result []terraform.DataSource
	}{
		{
			P:      &Provider{},
			Result: []terraform.DataSource{},
		},

		{
			P: &Provider{
				DataSourcesMap: map[string]*Resource{
					"foo": nil,
					"bar": nil,
				},
			},
			Result: []terraform.DataSource{
				terraform.DataSource{Name: "bar"},
				terraform.DataSource{Name: "foo"},
			},
		},
	}

	for i, tc := range cases {
		actual := tc.P.DataSources()
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d: got %#v; want %#v", i, actual, tc.Result)
		}
	}
}

func TestProviderValidate(t *testing.T) {
	cases := []struct {
		P      *Provider
		Config map[string]interface{}
		Err    bool
	}{
		{
			P: &Provider{
				Schema: map[string]*Schema{
					"foo": &Schema{},
				},
			},
			Config: nil,
			Err:    true,
		},
	}

	for i, tc := range cases {
		c, err := config.NewRawConfig(tc.Config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		_, es := tc.P.Validate(terraform.NewResourceConfig(c))
		if len(es) > 0 != tc.Err {
			t.Fatalf("%d: %#v", i, es)
		}
	}
}

func TestProviderValidateResource(t *testing.T) {
	cases := []struct {
		P      *Provider
		Type   string
		Config map[string]interface{}
		Err    bool
	}{
		{
			P:      &Provider{},
			Type:   "foo",
			Config: nil,
			Err:    true,
		},

		{
			P: &Provider{
				ResourcesMap: map[string]*Resource{
					"foo": &Resource{},
				},
			},
			Type:   "foo",
			Config: nil,
			Err:    false,
		},
	}

	for i, tc := range cases {
		c, err := config.NewRawConfig(tc.Config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		_, es := tc.P.ValidateResource(tc.Type, terraform.NewResourceConfig(c))
		if len(es) > 0 != tc.Err {
			t.Fatalf("%d: %#v", i, es)
		}
	}
}

func TestProviderImportState_default(t *testing.T) {
	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"foo": &Resource{
				Importer: &ResourceImporter{},
			},
		},
	}

	states, err := p.ImportState(&terraform.InstanceInfo{
		Type: "foo",
	}, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(states) != 1 {
		t.Fatalf("bad: %#v", states)
	}
	if states[0].ID != "bar" {
		t.Fatalf("bad: %#v", states)
	}
}

func TestProviderImportState_setsId(t *testing.T) {
	var val string
	stateFunc := func(d *ResourceData, meta interface{}) ([]*ResourceData, error) {
		val = d.Id()
		return []*ResourceData{d}, nil
	}

	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"foo": &Resource{
				Importer: &ResourceImporter{
					State: stateFunc,
				},
			},
		},
	}

	_, err := p.ImportState(&terraform.InstanceInfo{
		Type: "foo",
	}, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if val != "bar" {
		t.Fatal("should set id")
	}
}

func TestProviderImportState_setsType(t *testing.T) {
	var tVal string
	stateFunc := func(d *ResourceData, meta interface{}) ([]*ResourceData, error) {
		d.SetId("foo")
		tVal = d.State().Ephemeral.Type
		return []*ResourceData{d}, nil
	}

	p := &Provider{
		ResourcesMap: map[string]*Resource{
			"foo": &Resource{
				Importer: &ResourceImporter{
					State: stateFunc,
				},
			},
		},
	}

	_, err := p.ImportState(&terraform.InstanceInfo{
		Type: "foo",
	}, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if tVal != "foo" {
		t.Fatal("should set type")
	}
}

func TestProviderMeta(t *testing.T) {
	p := new(Provider)
	if v := p.Meta(); v != nil {
		t.Fatalf("bad: %#v", v)
	}

	expected := 42
	p.SetMeta(42)
	if v := p.Meta(); !reflect.DeepEqual(v, expected) {
		t.Fatalf("bad: %#v", v)
	}
}
