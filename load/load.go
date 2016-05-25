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

package load

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/asteris-llc/converge/resource"
)

// Load a module from a resource. This uses the protocol in the path (or file://
// if not present) to determine from where the module should be loaded.
func Load(source string) (*Graph, error) {
	initial, err := loadAny(nil, source)
	if err != nil {
		return nil, err
	}

	root, err := parseSource(source)
	if err != nil {
		return nil, err
	}

	modules := []*resource.Module{initial}
	for len(modules) > 0 {
		// bookkeeping to avoid recursive calls. Using `range` here would copy and
		// not process any new items.
		var module *resource.Module
		module, modules = modules[0], modules[1:]

		// actual work - loading modules recursively and replacing their ModuleTasks
		// with Modules
		for i, res := range module.Resources {
			if mt, ok := res.(*resource.ModuleTask); ok {
				newModule, err := loadAny(root, mt.Source)
				if err != nil {
					return nil, err
				}

				newModule.Args = mt.Args
				newModule.Source = mt.Source
				newModule.ModuleName = mt.ModuleName

				module.Resources[i] = newModule
				modules = append(modules, newModule)
			}
		}
	}

	return NewGraph(initial)
}

func parseSource(source string) (*url.URL, error) {
	url, err := url.Parse(source)
	if err != nil {
		return url, err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	return url, nil
}

func loadAny(root *url.URL, source string) (*resource.Module, error) {
	url, err := parseSource(source)
	if err != nil {
		return nil, err
	}

	if root != nil && !path.IsAbs(url.Path) {
		url.Path = path.Join(path.Dir(root.Path), url.Path)
	}

	switch url.Scheme {
	case "file":
		return FromFile(url.Path)

	default:
		return nil, fmt.Errorf("protocol %q is not implemented", url.Scheme)
	}
}

// FromFile loads a module from a file
func FromFile(filename string) (*resource.Module, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{"file", filename}
		}
		return nil, err
	}

	mod, err := Parse(content)

	if err == nil {
		mod.ModuleName = path.Base(filename)
	}

	return mod, err
}
