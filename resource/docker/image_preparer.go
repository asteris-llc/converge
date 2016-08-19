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
	"time"

	"github.com/asteris-llc/converge/resource"
)

// ImagePreparer for docker images
type ImagePreparer struct {
	Name    string `hcl:"name"`
	Tag     string `hcl:"tag"`
	Timeout string `hcl:"timeout"`
}

// Prepare a new docker image
func (p *ImagePreparer) Prepare(render resource.Renderer) (resource.Task, error) {
	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}

	tag, err := render.Render("tag", p.Tag)
	if err != nil {
		return nil, err
	}

	timeout, err := render.Render("timeout", p.Timeout)
	if err != nil {
		return nil, err
	}

	dockerClient, err := newDockerClient()
	if err != nil {
		return nil, err
	}

	if timeout != "" {
		duration, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
		dockerClient.PullInactivityTimeout = duration
	}

	image := &Image{
		Name: name,
		Tag:  tag,
	}
	image.SetClient(dockerClient)
	return image, nil
}
