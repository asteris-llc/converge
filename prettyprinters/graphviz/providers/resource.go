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

package printproviders

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/resource/module"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/template"
)

type ResourceProvider struct {
	graphviz.GraphIDProvider
}

func (p ResourceProvider) VertexGetLabel(e graphviz.GraphEntity) (string, error) {
	var name string

	if e.Name == "root" {
		name = "/"
	} else {
		name = strings.Split(e.Name, "root/")[1]
	}

	switch e.Value.(type) {
	case *template.Preparer:
		v := e.Value.(*template.Preparer)
		return fmt.Sprintf("Template: %s", v.Destination), nil
	case *module.Preparer:
		return fmt.Sprintf("Module: %s", name), nil
	case *param.Preparer:
		v := e.Value.(*param.Preparer)
		return fmt.Sprintf("%s = \\\"%s\\\"", name, v.Default), nil
	default:
		return name, nil
	}
}

func (p ResourceProvider) VertexGetProperties(e graphviz.GraphEntity) graphviz.PropertySet {
	properties := make(map[string]string)
	if e.Name == "root" {
		properties["style"] = "invis"
	}
	switch e.Value.(type) {
	case *shell.Preparer:
		properties["shape"] = "component"
	case *template.Preparer:
		properties["shape"] = "tab"
	}
	return properties
}

func (p ResourceProvider) EdgeGetProperties(src graphviz.GraphEntity, dst graphviz.GraphEntity) graphviz.PropertySet {
	properties := make(map[string]string)
	if src.Name == "root" {
		properties["style"] = "invis"
	}
	return properties
}

func (p ResourceProvider) SubgraphMarker(e graphviz.GraphEntity) graphviz.SubgraphMarkerKey {
	switch e.Value.(type) {
	case *module.Preparer:
		return graphviz.SubgraphMarkerStart
	default:
		return graphviz.SubgraphMarkerNOP
	}
}

func ResourcePreparer() graphviz.GraphvizPrintProvider {
	return ResourceProvider{}
}
