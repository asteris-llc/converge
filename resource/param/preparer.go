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
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

const (
	PASS int = iota
	WARN
	FAIL
	OTHER_ERR
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
	// right now, it can either be "int" or "string" (the default). This
	// is used with TypeCastValue.
	Type string `hcl:"type"`

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

	typedValue, err := TypeCastValue(value, p.Type)
	if err != nil {
		return nil, err
	}

	returnCode, err := TypeValidation(typedValue, p.Predicates())
	if err != nil {
		return nil, err
	}

	if returnCode != 0 {
		return nil, fmt.Errorf("param failed validation, with return code %d", returnCode)
	}

	return &Param{Value: value}, nil
}

func (p *Preparer) Predicates() map[string]string {

	predicates := make(map[string]string)

	closure := func(rule string, returnCode int) {
		count := len(predicates)
		name := fmt.Sprintf("pred#%d", count)
		predicate := fmt.Sprintf("{{if %s }}0{{else}}%d{{end}}", rule, returnCode)
		predicates[name] = predicate
	}

	for _, rule := range p.Must {
		closure(rule, FAIL)
	}

	return predicates
}

func TypeCastValue(paramValue, paramType string) (interface{}, error) {
	switch paramType {
	case "int":
		value, err := strconv.Atoi(paramValue)
		if err != nil {
			return value, fmt.Errorf("paramType is \"int\", but converting \"%s\" failed", paramValue)
		}
		return value, nil

	// this case is a nop, since string is the default type for parameters
	case "string":
		return paramValue, nil
	case "":
		// try to infer the type from the value
		// the recursion is make the code DRYer, but is this the best way?
		if value, err := TypeCastValue(paramValue, "int"); err == nil {
			return value, nil
		} else {
			return TypeCastValue(paramValue, "string")
		}
	default:
		return paramValue, fmt.Errorf("%s is not a supported param type", paramType)
	}
}

func TypeValidation(value interface{}, predicates map[string]string) (int, error) {
	// nop if no predicates to test
	// this reduces boilerplate for callers
	if len(predicates) == 0 {
		return PASS, nil
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
		oneOfClosure := func(list ...string) bool {
			for _, item := range list {
				if str == item {
					return true
				}
			}
			return false
		}

		funcMap = template.FuncMap{
			"empty":    func() bool { return len(str) == 0 },
			"notEmpty": func() bool { return len(str) != 0 },
			"oneOf":    oneOfClosure,
			"notOneOf": func(list ...string) bool { return !oneOfClosure(list...) },
		}
	}

	for name, predicate := range predicates {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(predicate)
		if err != nil {
			return OTHER_ERR, err
		}

		var buffer bytes.Buffer
		if err = tmpl.Execute(&buffer, value); err != nil {
			return OTHER_ERR, err
		}

		returnCode, err := strconv.Atoi(buffer.String())
		if err != nil {
			return OTHER_ERR, err
		}

		if returnCode != 0 {
			log.Debug(fmt.Sprintf("%s", predicate))
			return returnCode, fmt.Errorf("%s: expected 0, got %d", tmpl.Name(), returnCode)
		}
	}
	return PASS, nil
}

func init() {
	registry.Register("param", (*Preparer)(nil), (*Param)(nil))
}
