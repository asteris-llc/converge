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

package param

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"text/template"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for params
//
// Param controls the flow of values through `module` calls. You can use the
// `{{param "name"}}` template call anywhere you need the value of a param
// inside the current module.
type Preparer struct {
	// Default is an optional field that provides a default value if none is
	// provided to this parameter. If this field is not set, this param will be
	// treated as required.
	Default *string `hcl:"default"`
	Type    string  `hcl:"type"`
	Rules   []*Rule `hcl:"rule"`
}

type Rule struct {
	Description string   `hcl:"description"`
	Must        []string `hcl:"must"`
	Should      []string `hcl:"should"`
}

type ParamTest struct {
	level    int
	template *template.Template
	value    interface{}
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	val, present := render.Value()

	if !present {
		if p.Default == nil {
			return nil, errors.New("param is required")
		} else {
			def, err := render.Render("default", *p.Default)
			if err != nil {
				return nil, err
			}

			if _, err := strconv.Atoi(def); err == nil {
				p.Type = "int"
			} else {
				p.Type = "string"
			}

			val = def
		}
	} else {
		if p.Type == "" {
			p.Type = "string"
		}
	}

	var paramTests []*ParamTest
	lang, err := ValidationLanguage(val, p.Type)
	if err != nil {
		return nil, err
	}

	for rNum, rule := range p.Rules {
		pt := &ParamTest{value: val}
		for mNum, must := range rule.Must {
			must = "{{" + must + "}}"
			pt.level = 1
			name := fmt.Sprintf("rule-%d-must-%d", rNum, mNum)
			pt.template, err = template.New(name).Funcs(lang.Funcs).Parse(must)
			if err != nil {
				return nil, err
			}
			paramTests = append(paramTests, pt)
		}
		for sNum, should := range rule.Should {
			should = "{{" + should + "}}"
			pt.level = 2
			name := fmt.Sprintf("rule-%d-should-%d", rNum, sNum)
			pt.template, err = template.New(name).Funcs(lang.Funcs).Parse(should)
			if err != nil {
				return nil, err
			}
			paramTests = append(paramTests, pt)
		}
	}

	for _, pt := range paramTests {
		if err = pt.Validate(); err != nil {
			return nil, err
		}
	}

	return &Param{Value: val}, nil
}

func (pt *ParamTest) Validate() error {
	var buffer bytes.Buffer
	if err := pt.template.Execute(&buffer, pt.value); err != nil {
		return err
	}

	if pass := buffer.String(); pass != "true" {
		return fmt.Errorf("Expected true from %s, got %s", pt.template.Name(), pass)
	}

	return nil
}

func ValidationLanguage(val, vtype string) (*extensions.LanguageExtension, error) {
	lang := extensions.DefaultLanguage()

	switch vtype {
	case "int":
		num, err := strconv.Atoi(val)
		if err != nil {
			return lang, fmt.Errorf("vtype is \"int\", but converting \"%s\" failed", val)
		}
		lang.On("min", func(min int) bool { return num >= min })
		lang.On("max", func(max int) bool { return num <= max })
		lang.On("range", func(min, max int) bool { return min <= num && num <= max })
		lang.On("isEven", func() bool { return math.Mod(float64(num), 2) == 0 })
		lang.On("isOdd", func() bool { return math.Mod(float64(num), 2) != 0 })
	case "string":
		lang.On("empty", func() bool { return len(val) == 0 })
		lang.On("oneOf", func(list ...string) bool {
			isOneOf := false
			for _, item := range list {
				if val == item {
					isOneOf = true
				}
			}
			return isOneOf
		})
		lang.On("notIn", func(list ...string) bool {
			for _, item := range list {
				if val == item {
					return false
				}
			}
			return true
		})
	default:
		return lang, fmt.Errorf("Unhandled case %T for %v", vtype, val)
	}

	return lang, nil
}

func init() {
	registry.Register("param", (*Preparer)(nil), (*Param)(nil))
}
