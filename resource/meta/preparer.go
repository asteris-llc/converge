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

package meta

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

type Preparer struct {
	Author       string `hcl:"author"`
	Organization string `hcl:"organization"`
	PgpKeyId     string `hcl:"pgp_key_id"`
	OrgUrl       string `hcl:"org_url"`
	Version      string `hcl:"version"`
	VcsUrl       string `hcl:"vcs_url"`
	License      string `hcl:"license"`
	VcsCommit    string `hcl:"vcs_commit"`
	Description  string `hcl:"description"`
	//Platforms    []map[string]interface{} `hcl:"platforms"`
}

func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// get the fields from the struct, then String them
	metaValue := reflect.ValueOf(m).Elem()
	stringSlice := []string{"meta:"}

	for i := 0; i < metaValue.NumField(); i++ {
		key := metaValue.Type().Field(i).Name
		value := metaValue.Field(i)

		stringSlice = append(stringSlice, fmt.Sprintf("%v:\t%v", key, value))

	}
	return nil, errors.New("Not implemented yet")
}

func init() {
	registry.Register("meta", (*Preparer)(nil), (*Meta)(nil))
}
