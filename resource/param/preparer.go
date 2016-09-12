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
	"fmt"
	"strconv"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

type ParamType string

const (
	ParamTypeString   ParamType = "string"
	ParamTypeInt      ParamType = "int"
	ParamTypeInferred ParamType = ""
)

const (
	ValidatePass int = 0
	ValidateFail int = 2
)

// Rule for Type Validation
//
// Rule holds a predicate for parameter validation. It takes a text/template
// fragment as input. The validation logic then wraps the expression around a
// different template fragment, then evals it to determine the outcome

// Preparer for params
//
// Param controls the flow of values through `module` calls. You can use the
// `{{param "name"}}` template call anywhere you need the value of a param
// inside the current module.
type Preparer struct {
	// Default is an optional field that provides a default value if none
	// is provided to this parameter. If this field is not set, this param
	// will be treated as required. If this is provided, then Type is
	// inferred
	Default *string `hcl:"default"`

	// Type is an optional field that is used for type checking. As of
	// right now, it can either be ParamTypeString or ParamTypeInt. It
	// is used with typeCastValue.
	Type ParamType `hcl:"type"`

	Must []string `hcl:"must"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	value, present := render.Value()

	if !present {
		if p.Default == nil {
			return nil, fmt.Errorf("param is required")
		}
		def, err := render.Render("default", *p.Default)
		if err != nil {
			return nil, err
		}
		value = def
	}

	typedValue, err := typeCastValue(value, p.Type)
	if err != nil {
		return nil, err
	}

	err = ValidateType(typedValue, p.Predicates())
	if err != nil {
		return nil, err
	}

	return &Param{Value: value}, nil
}

func (p *Preparer) Predicates() map[string]string {

	predicates := make(map[string]string)

	for _, rule := range p.Must {
		count := len(predicates)
		name := fmt.Sprintf("pred#%d", count)
		predicate := fmt.Sprintf("{{if %s }}%d{{else}}%d{{end}}", rule, ValidatePass, ValidateFail)
		predicates[name] = predicate
	}

	return predicates
}

func typeCastValue(paramValue string, paramType ParamType) (interface{}, error) {
	switch paramType {
	case ParamTypeInt:
		value, err := strconv.Atoi(paramValue)
		if err != nil {
			return value, fmt.Errorf("paramType is \"int\", but converting \"%s\" failed", paramValue)
		}
		return value, nil

	// this case is a nop, since string is the default type for parameters
	case ParamTypeString:
		return paramValue, nil
	case ParamTypeInferred:
		// try to infer the type from the value
		// the recursion is make the code DRYer, but is this the best way?
		if value, err := typeCastValue(paramValue, ParamTypeInt); err == nil {
			return value, nil
		} else {
			return typeCastValue(paramValue, ParamTypeString)
		}
	default:
		return paramValue, fmt.Errorf("%s is not a supported param type", paramType)
	}
}

func ValidateType(value interface{}, predicates map[string]string) error {
	// nop if no predicates to test
	// this reduces boilerplate for callers
	if len(predicates) == 0 {
		return nil
	}

	var funcMap template.FuncMap
	switch value.(type) {
	case int:
		num := value.(int)
		funcMap = template.FuncMap{
			"min":    func(min int) bool { return num >= min },
			"max":    func(max int) bool { return num <= max },
			"isEven": func() bool { return num%2 == 0 },
			"isOdd":  func() bool { return num%2 != 0 },
		}

	case string:
		str := value.(string)

		funcMap = template.FuncMap{
			"empty":    func() bool { return len(str) == 0 },
			"notEmpty": func() bool { return len(str) != 0 },
			"oneOf": func(list string) bool {
				for _, item := range strings.Split(list, " ") {
					if str == item {
						return true
					}
				}
				return false
			},
			"notOneOf": func(list string) bool {
				return !funcMap["oneOf"].(func(list string) bool)(list)
			},
		}
	}

	for name, predicate := range predicates {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(predicate)
		if err != nil {
			return err
		}

		var buffer bytes.Buffer
		if err = tmpl.Execute(&buffer, value); err != nil {
			return err
		}

		returnCode, err := strconv.Atoi(buffer.String())
		if err != nil {
			return err
		}

		if returnCode != 0 {
			log.Debug(fmt.Sprintf("%s", predicate))
			return fmt.Errorf("%s: expected %d, got %d", tmpl.Name(), ValidatePass, returnCode)
		}
	}
	return nil
}

func init() {
	registry.Register("param", (*Preparer)(nil), (*Param)(nil))
}
