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
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/asteris-llc/converge/fetch"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/keystore"
	"github.com/asteris-llc/converge/parse"
	"github.com/pkg/errors"
)

type source struct {
	Parent       string
	ParentSource string
	Source       string
}

func (s *source) String() string {
	return fmt.Sprintf("%s (%s)", s.Source, s.Parent)
}

// Nodes loads and parses all resources referred to by the provided url
func Nodes(ctx context.Context, root string, verify bool) (*graph.Graph, error) {
	toLoad := []*source{{"root", root, root}}

	out := graph.New()
	out.Add("root", nil)

	for len(toLoad) > 0 {
		select {
		case <-ctx.Done():
			return nil, errors.New("interrupted")
		default:
		}

		current := toLoad[0]
		toLoad = toLoad[1:]

		url, err := fetch.ResolveInContext(current.Source, current.ParentSource)
		if err != nil {
			return nil, err
		}

		log.Printf("[DEBUG] fetching %s\n", url)
		content, err := fetch.Any(ctx, url)
		if err != nil {
			return nil, errors.Wrap(err, url)
		}

		if verify {
			signatureURL := url + ".asc"

			log.Printf("[DEBUG] fetching %s\n", signatureURL)
			signature, err := fetch.Any(ctx, signatureURL)
			if err != nil {
				return nil, errors.Wrap(err, signatureURL)
			}

			ks := keystore.Default()
			err = ks.CheckSignature(bytes.NewBuffer(content), bytes.NewBuffer(signature))

			if err != nil {
				return nil, errors.Wrap(err, signatureURL)
			}
		}

		resources, err := parse.Parse(content)
		if err != nil {
			return nil, errors.Wrap(err, url)
		}

		for _, resource := range resources {
			newID := graph.ID(current.Parent, resource.String())
			out.Add(newID, resource)
			out.ConnectParent(current.Parent, newID)

			if resource.IsModule() {
				toLoad = append(
					toLoad,
					&source{
						Parent:       newID,
						ParentSource: url,
						Source:       resource.Source(),
					},
				)
			}
		}
	}

	return out, out.Validate()
}
