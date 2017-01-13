// Copyright © 2017 Asteris, LLC
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

package unarchive

import (
	"net/url"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Preparer for Unarchive
//
// Unarchive renders file unarchive data
type Preparer struct {
	// Source to unarchive
	Source string `hcl:"source" required:"true" nonempty:"true"`

	// Destination for the unarchived file
	Destination string `hcl:"destination" required:"true" nonempty:"true"`
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if strings.TrimSpace(p.Source) == "" {
		return nil, errors.New("\"source\" must contain a value")
	}

	_, err := url.Parse(p.Source)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse \"source\"")
	}

	if strings.TrimSpace(p.Destination) == "" {
		return nil, errors.New("\"destination\" must contain a value")
	}

	return &Unarchive{
		Source:      p.Source,
		Destination: p.Destination,
	}, nil
}

func init() {
	registry.Register("file.unarchive", (*Preparer)(nil), (*Unarchive)(nil))
}
