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

package docker

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
)

// Image is responsible for pulling docker images
type Image struct {
	Name       string
	Tag        string
	client     APIClient
	diffs      map[string]resource.Diff
	statusCode int
}

// Check system for presence of docker image
func (i *Image) Check() (resource.TaskStatus, error) {
	repoTag := i.RepoTag()
	i.diffs = make(map[string]resource.Diff)
	imagesDiff := resource.TextDiff{Values: [2]string{"<image-missing>", repoTag}}

	image, err := i.client.FindImage(repoTag)
	if err != nil {
		i.statusCode = resource.StatusFatal
		return i, err
	}

	if image != nil {
		imagesDiff.Values[0] = repoTag
	}

	i.diffs["image"] = imagesDiff
	return i, nil
}

// Apply pulls a docker image
func (i *Image) Apply() (err error) {
	return i.client.PullImage(i.Name, i.Tag)
}

// Value returns the repo tag of the desired docker image
func (i *Image) Value() string {
	return i.RepoTag()
}

// Diffs returns the diff map for the docker image
func (i *Image) Diffs() map[string]resource.Diff {
	return i.diffs
}

// StatusCode returns the status code
func (i *Image) StatusCode() int {
	return i.statusCode
}

// Messages returns status messages for the docker image resource
func (i *Image) Messages() []string {
	return nil
}

// Changes returns true if a docker image needs to be pulled
func (i *Image) Changes() bool {
	return resource.AnyChanges(i.diffs)
}

// SetClient injects a docker api client
func (i *Image) SetClient(client APIClient) {
	i.client = client
}

// RepoTag builds a repo tag used to identify a specific docker image
func (i *Image) RepoTag() string {
	var tag string
	if i.Tag != "" {
		tag = i.Tag
	} else {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", i.Name, tag)
}
