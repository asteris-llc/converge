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

package container

import (
	"fmt"

	"github.com/asteris-llc/converge/helpers/transform"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	"golang.org/x/net/context"
)

// Preparer for docker containers
//
// Container is responsible for creating docker containers. It assumes that
// there is already a Docker daemon running on the system.
// *Note: docker resources are not currently supported on Solaris.*
type Preparer struct {
	// name of the container
	Name string `hcl:"name" required:"true" nonempty:"true"`

	// the image name or ID to use for the container
	Image string `hcl:"image" required:"true" nonempty:"true"`

	// override the container entrypoint
	Entrypoint []string `hcl:"entrypoint"`

	// override the container command
	Command []string `hcl:"command"`

	// override the working directory of the container
	WorkingDir string `hcl:"working_dir" nonempty:"true"`

	// set environment variables in the container
	Env map[string]string `hcl:"env"`

	// additional ports to expose in the container
	Expose []string `hcl:"expose"`

	// A list of links for the container. Each link entry should be in the form of
	// container_name:alias
	Links []string `hcl:"links"`

	// publish container ports to the host. Each item should be in the following
	// format:
	// ip:hostPort:containerPort|ip::containerPort|hostPort:containerPort|containerPort.
	// Ports can be specified in the format: portnum/proto. If proto is not
	// specified, "tcp" is assumed
	Ports []string `hcl:"ports"`

	// list of DNS servers for the container to use
	DNS []string `hcl:"dns"`

	// the mode of the container network. default: default
	NetworkMode string `hcl:"network_mode"`

	// the networks to connect the container to
	Networks []string `hcl:"networks"`

	// bind mounts volumes
	Volumes []string `hcl:"volumes"`

	// mounts all volumes from the specified container
	VolumesFrom []string `hcl:"volumes_from"`

	// allocates a random host port for all of a container’s exposed ports.
	// Specified as a boolean value
	PublishAllPorts bool `hcl:"publish_all_ports"`

	// the desired status of the container.
	Status string `hcl:"status" valid_values:"running,created"`

	// indicates whether or not the container will be recreated if the state is
	// not what is expected. By default, the module will only check to see if the
	// container exists. Specified as a boolean value
	Force bool `hcl:"force"`
}

// Prepare a docker container
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	env := transform.StringsMapToStringSlice(
		p.Env,
		func(k, v string) string {
			return fmt.Sprintf("%s=%s", k, v)
		},
	)

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	if p.NetworkMode == "" {
		p.NetworkMode = DefaultNetworkMode
	}

	container := &Container{
		Force:           p.Force,
		Name:            p.Name,
		CStatus:         p.Status,
		Image:           p.Image,
		Entrypoint:      p.Entrypoint,
		Command:         p.Command,
		WorkingDir:      p.WorkingDir,
		Env:             env,
		Expose:          p.Expose,
		Links:           p.Links,
		NetworkMode:     p.NetworkMode,
		Networks:        p.Networks,
		PublishAllPorts: p.PublishAllPorts,
		PortBindings:    p.Ports,
		DNS:             p.DNS,
		Volumes:         p.Volumes,
		VolumesFrom:     p.VolumesFrom,
	}
	container.SetClient(dockerClient)
	return container, nil
}

func init() {
	registry.Register("docker.container", (*Preparer)(nil), (*Container)(nil))
}
