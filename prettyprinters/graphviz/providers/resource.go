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

package printers

import (
	"fmt"

	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/template"
)

type ResourceProvider struct{}

func getResourceName(res resource.Resource) (string, error) {
	switch res.(type) {
	case *module.Preparer:
	case *param.Preparer:
	case *shell.Preparer:
	case *template.Preparer:
	default:
		return fmt.Sprintf("%T", res), nil
	}
	return "", nil
}

func (p ResourceProvider) VertexGetID(res interface{}) (string, error) {
	return "", nil
}

func (p ResourceProvider) VertexGetLabel(res interface{}) (string, error) {
	return "", nil
}

func (p ResourceProvider) VertexGetProperties(res interface{}) graphviz.PropertySet {
	return make(graphviz.PropertySet)
}

func (p ResourceProvider) EdgeGetLabel(srcRes interface{}, destRes interface{}) (string, error) {
	return "", nil
}

func (p ResourceProvider) EdgeGetProperties(srcRes interface{}, destRes interface{}) graphviz.PropertySet {
	return make(graphviz.PropertySet)
}

func (p ResourceProvider) SubgraphMarker(res interface{}) graphviz.SubgraphMarkerKey {
	return graphviz.SubgraphMarkerNOP
}
