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

package resource

// DependencyTracker is an embedded behavior for tracking dependencies.
type DependencyTracker struct {
	BaseItems     *[]string `hcl:"depends"`
	ComputedItems []string
}

// SetDepends sets the list of provided dependencies
func (dt *DependencyTracker) SetDepends(items []string) {
	dt.BaseItems = &items
}

// HasBaseDependencies indicates if BaseItems is set
func (dt *DependencyTracker) HasBaseDependencies() bool {
	return dt.BaseItems != nil
}

// Depends lists tracked dependencies
func (dt *DependencyTracker) Depends() []string {
	var base []string
	if dt.BaseItems != nil {
		base = *dt.BaseItems
	}

	return dt.dedupe(
		base,
		dt.ComputedItems,
	)
}

// ComputeDependencies for the given strings
func (dt *DependencyTracker) ComputeDependencies(name string, renderer *Renderer, sources ...string) error {
	results := [][]string{}

	for _, source := range sources {
		result, err := renderer.Params(name, source)
		if err != nil {
			return err
		}

		results = append(results, result)
	}

	dt.ComputedItems = dt.dedupe(results...)
	return nil
}

func (dt *DependencyTracker) dedupe(sources ...[]string) (out []string) {
	dest := map[string]struct{}{}

	for _, source := range sources {
		for _, item := range source {
			dest[item] = struct{}{}
		}
	}

	for k := range dest {
		out = append(out, k)
	}

	return out
}
