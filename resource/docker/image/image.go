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

// +build !solaris

package image

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	"golang.org/x/net/context"
)

// Image is responsible for pulling docker images
type Image struct {
	Name   string `export:"name"`
	Tag    string `export:"tag"`
	client docker.APIClient
}

// Check system for presence of docker image
func (i *Image) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	repoTag := i.RepoTag()
	image, err := i.client.FindImage(repoTag)
	if err != nil {
		status.Level = resource.StatusFatal
		return status, err
	}

	var original string
	if image != nil {
		original = repoTag
	}

	status.AddDifference("image", original, repoTag, "<image-missing>")
	if resource.AnyChanges(status.Differences) {
		status.Level = resource.StatusWillChange
	}
	return status, nil
}

// Apply pulls a docker image
func (i *Image) Apply(context.Context) (resource.TaskStatus, error) {
	if err := i.client.PullImage(i.Name, i.Tag); err != nil {
		return &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}, err
	}
	return &resource.Status{}, nil
}

// SetClient injects a docker api client
func (i *Image) SetClient(client docker.APIClient) {
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
