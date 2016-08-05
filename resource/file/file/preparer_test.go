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

package file_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/file"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(file.Preparer))
}

func TestVaildPreparer(t *testing.T) {
	t.Parallel()
	preparers := []resource.Resource{
		//Test default state
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody"},
		//Test explicit file state
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSFile)},
		//Test absent state
		&file.Preparer{Destination: "/path/to/file", State: string(file.FSAbsent)},
		//Test Touch
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSTouch)},
		//Link
		&file.Preparer{Source: "/path/to/other/file", Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSLink)},
		//Hard
		&file.Preparer{Source: "/path/to/other/file", Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSHard)},
		//Directory
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSDirectory)},
		//Recurse
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSDirectory), Recurse: "true"},
	}
	errs := []string{
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}
	helpers.PreparerValidator(t, preparers, errs)
}

func TestInVaildPreparer(t *testing.T) {
	t.Parallel()
	preparers := []resource.Resource{
		//Test useless file module
		&file.Preparer{Destination: "/path/to/file"},
		//Test useless file module
		&file.Preparer{Destination: "/path/to/file", State: string(file.FSFile)},
		//Test absent state
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSAbsent)},
		//Link
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSLink)},
		//Hard
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", State: string(file.FSHard)},
		//Recurse
		&file.Preparer{Destination: "/path/to/file", Mode: "0777", User: "nobody", Recurse: "true"},
	}
	errs := []string{
		"useless file module",
		"useless file module",
		"cannot use `mode` or `owner` parameters when `state` is set to \"absent\"",
		"module 'file' requires a `source` parameter when `state`=\"link\"",
		"module 'file' requires a `source` parameter when `state`=\"hard\"",
		"cannot use the `recurse` parameter when `state` is not set to \"directory\"",
	}
	helpers.PreparerValidator(t, preparers, errs)
}
