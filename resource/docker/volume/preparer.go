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

package volume

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	"golang.org/x/net/context"
)

// Preparer for docker volumes
//
// Volume is responsible for managing Docker volumes. It assumes that there is
// already a Docker daemon running on the system.
// *Note: docker resources are not currently supported on Solaris.*
type Preparer struct {
	// name of the volume
	Name string `hcl:"name" required:"true" nonempty:"true"`

	// volume driver. default: local
	Driver string `hcl:"driver" default:"local"`

	// labels to set on the volume
	Labels map[string]string `hcl:"labels"`

	// driver specific options
	Options map[string]string `hcl:"options"`

	// indicates whether the volume should exist.
	State State `hcl:"state" valid_values:"present,absent"`

	// indicates whether or not the volume will be recreated if the state is not
	// what is expected. By default, the module will only check to see if the
	// volume exists. Specified as a boolean value
	Force bool `hcl:"force"`
}

// Prepare a docker volume
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	state := p.State
	if state == "" {
		state = StatePresent
	}

	driver := p.Driver
	if driver == "" {
		driver = "local"
	}

	volume := &Volume{
		Name:    p.Name,
		Driver:  driver,
		Labels:  p.Labels,
		Options: p.Options,
		State:   state,
		Force:   p.Force,
	}
	volume.SetClient(dockerClient)
	return volume, nil
}

func init() {
	registry.Register("docker.volume", (*Preparer)(nil), (*Volume)(nil))
}
