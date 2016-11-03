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

package network

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
)

// Preparer for docker networks
//
// Network is responsible for managing Docker networks. It assumes that there is
// already a Docker daemon running on the system.
type Preparer struct {
	// name of the network
	Name string `hcl:"name" required:"true"`

	// network driver. default: bridge
	Driver string `hcl:"driver"`

	// labels to set on the network
	Labels map[string]string `hcl:"labels"`

	// driver specific options
	Options map[string]interface{} `hcl:"options"`

	// indicates whether the volume should exist.
	State State `hcl:"state" valid_values:"present,absent"`

	// indicates whether or not the volume will be recreated if the state is not
	// what is expected. By default, the module will only check to see if the
	// volume exists. Specified as a boolean value
	Force bool `hcl:"force"`
}

// Prepare a docker network
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	if p.Driver == "" {
		p.Driver = DefaultDriver
	}

	if p.State == "" {
		p.State = "present"
	}

	nw := &Network{
		Name:    p.Name,
		Driver:  p.Driver,
		Labels:  p.Labels,
		Options: p.Options,
		State:   p.State,
		Force:   p.Force,
	}
	nw.SetClient(dockerClient)
	return nw, nil
}

func init() {
	registry.Register("docker.network", (*Preparer)(nil), (*Network)(nil))
}
