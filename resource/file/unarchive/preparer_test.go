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

package unarchive_test

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/unarchive"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// TestPreparerInterface tests that the Preparer interface is properly implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(unarchive.Preparer))
}

// TestPreparer tests the valid and invalid cases of Prepare
func TestPreparer(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}

	t.Run("valid", func(t *testing.T) {
		p := unarchive.Preparer{
			Source:      "/tmp/test.zip",
			Destination: "/tmp/test",
		}
		_, err := p.Prepare(context.Background(), &fr)
		assert.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("source", func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				p := unarchive.Preparer{
					Source:      "",
					Destination: "/tmp/test",
				}
				_, err := p.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"source\" must contain a value"))
			})

			t.Run("space", func(t *testing.T) {
				p := unarchive.Preparer{
					Source:      " ",
					Destination: "/tmp/test",
				}
				_, err := p.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"source\" must contain a value"))
			})

			t.Run("cannot parse", func(t *testing.T) {
				p := unarchive.Preparer{
					Source:      ":test",
					Destination: "/tmp/test",
				}
				_, err := p.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("failed to parse \"source\": parse %s: missing protocol scheme", p.Source))
			})
		})

		t.Run("destination", func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				p := unarchive.Preparer{
					Source:      "/tmp/test.zip",
					Destination: "",
				}
				_, err := p.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"destination\" must contain a value"))
			})

			t.Run("space", func(t *testing.T) {
				p := unarchive.Preparer{
					Source:      "/tmp/test.zip",
					Destination: " ",
				}
				_, err := p.Prepare(context.Background(), &fr)
				assert.EqualError(t, err, fmt.Sprintf("\"destination\" must contain a value"))
			})
		})
	})
}
