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

package render

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/resource/module"
)

// Factory generates Renderers
type Factory struct {
	Graph     *graph.Graph
	DotValues map[string]*LazyValue
	Language  *extensions.LanguageExtension
}

// ValueThunk lazily evaluates a param
type ValueThunk func() (string, bool, error)

// LazyValue wraps a ValueThunk in an interface and provides a way to cache the
// thunk evaluation.
type LazyValue struct {
	val interface{}
}

// Value returns the result of a ValueThunk, evaluating it if necessary
func (v *LazyValue) Value() (string, bool, error) {
	switch result := v.val.(type) {
	case [3]interface{}:
		return result[0].(string), result[1].(bool), result[2].(error)
	case ValueThunk:
		val, found, err := result()
		v.val = [3]interface{}{val, found, err}
		return val, found, err
	default:
		return "", false, errors.New("value is not a thunk")
	}
}

// GetRenderer returns a Factory for the specific graph node
func (f *Factory) GetRenderer(id string) (*Renderer, error) {
	fmt.Println("getting renderer for: ", id)
	r := &Renderer{Language: f.Language, Graph: func() *graph.Graph { return f.Graph }, ID: id}
	if dotVal, found := f.DotValues[id]; found {
		if valResult, valFound, err := dotVal.Value(); err != nil {
			return nil, err
		} else if valFound {
			r.DotValue = valResult
			r.DotValuePresent = true
		}
	}
	return r, nil
}

// HandleLookup performs a cross-node lookup
func (f *Factory) HandleLookup(id string) interface{} {
	fmt.Println("factory lookup for " + id)
	return nil
}

// NewFactory generates a new Render factory
func NewFactory(ctx context.Context, g *graph.Graph) (*Factory, error) {
	f := &Factory{
		Graph:     g,
		Language:  extensions.DefaultLanguage(),
		DotValues: make(map[string]*LazyValue),
	}

	f.Language = f.Language.On("lookup", f.HandleLookup)
	for _, vertex := range g.Vertices() {
		if dotVal, found := getParamOverrides(func() *graph.Graph { return f.Graph }, vertex); found {
			f.DotValues[vertex] = &LazyValue{dotVal}
		}
	}
	return f, nil
}

func getParamOverrides(gFunc func() *graph.Graph, id string) (ValueThunk, bool) {
	name := graph.BaseID(id)
	f := func() (string, bool, error) { return "", false, nil }
	if strings.HasPrefix(name, "param") {
		f = func() (string, bool, error) {
			fmt.Println("getting overrides for param: ", name)
			parent, ok := gFunc().GetParent(id).(*module.Module)
			if !ok {
				p := gFunc().GetParent(id)
				return "", false, fmt.Errorf("Parent of param %s was not a module, was %s :: %T", id, p, p)
			}
			if val, ok := parent.Params[name[len("param."):]]; ok {
				fmt.Println("found overrides")
				return val, true, nil
			}
			return "", false, nil
		}
	}
	return f, true
}
