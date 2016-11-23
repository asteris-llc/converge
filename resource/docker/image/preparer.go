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
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	"golang.org/x/net/context"
)

// Preparer for docker images
//
// Image is responsible for pulling Docker images. It assumes that there is
// already a Docker daemon running on the system.
type Preparer struct {
	// name of the image to pull
	Name string `hcl:"name" required:"true" nonempty:"true"`

	// tag of the image to pull. default: latest
	Tag string `hcl:"tag"`

	// the amount of time to wait after a period of inactivity. The timeout is
	// reset each time new data arrives.
	InactivityTimeout time.Duration `hcl:"inactivity_timeout"`
}

// Prepare a new docker image
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	dockerClient.PullInactivityTimeout = p.InactivityTimeout

	image := &Image{
		Name: p.Name,
		Tag:  p.Tag,
	}
	image.SetClient(dockerClient)
	return image, nil
}

func init() {
	registry.Register("docker.image", (*Preparer)(nil), (*Image)(nil))
}
