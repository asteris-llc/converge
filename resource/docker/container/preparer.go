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

package container

import (
	"errors"
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
)

// Preparer for docker containers
type Preparer struct {
	// name of the container
	Name string `hcl:"name"`

	// the image name or ID to use for the container
	Image string `hcl:"image"`

	// override the container entrypoint
	Entrypoint []string `hcl:"entrypoint"`

	// override the container command
	Command []string `hcl:"command"`

	// override the working directory of the container
	WorkingDir string `hcl:"working_dir"`

	// set environmnet variables in the container
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

	// bind mounts volumes
	Volumes []string `hcl:"volumes"`

	// mounts all volumes from the specified container
	VolumesFrom []string `hcl:"volumes_from"`

	// allocates a random host port for all of a container’s exposed ports.
	// Specified as a boolean value
	PublishAllPorts string `hcl:"publish_all_ports" doc_type:"bool"`

	// the desired status of the container. running|created
	Status string `hcl:"status"`

	// indicates whether or not the container will be recreated if the state is
	// not what is expected. By default, the module will only check to see if the
	// container exists. Specified as a boolean value
	Force string `hcl:"force" doc_type:"bool"`
}

// Prepare a docker container
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	force, err := render.RenderBool("force", p.Force)
	if err != nil {
		return nil, err
	}

	name, err := render.RequiredRender("name", p.Name)
	if err != nil {
		return nil, err
	}

	image, err := render.RequiredRender("image", p.Image)
	if err != nil {
		return nil, err
	}

	renderedEntrypoint, err := render.RenderStringSlice("entrypoint", p.Entrypoint)
	if err != nil {
		return nil, err
	}

	renderedCommand, err := render.RenderStringSlice("command", p.Command)
	if err != nil {
		return nil, err
	}

	workDir, err := render.Render("working_dir", p.WorkingDir)
	if err != nil {
		return nil, err
	}

	status, err := render.Render("status", p.Status)
	if err != nil {
		return nil, err
	}

	renderedEnv, err := render.RenderStringMapToStringSlice("env", p.Env, func(k, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	})
	if err != nil {
		return nil, err
	}

	renderedExpose, err := render.RenderStringSlice("expose", p.Expose)
	if err != nil {
		return nil, err
	}

	renderedLinks, err := render.RenderStringSlice("links", p.Links)
	if err != nil {
		return nil, err
	}

	publishAllPorts, err := render.RenderBool("publish_all_ports", p.PublishAllPorts)
	if err != nil {
		return nil, err
	}

	renderedPorts, err := render.RenderStringSlice("ports", p.Ports)
	if err != nil {
		return nil, err
	}

	renderedDNS, err := render.RenderStringSlice("dns", p.DNS)
	if err != nil {
		return nil, err
	}

	renderedVolumes, err := render.RenderStringSlice("volumes", p.Volumes)
	if err != nil {
		return nil, err
	}

	renderedVolumesFrom, err := render.RenderStringSlice("volumes_from", p.VolumesFrom)
	if err != nil {
		return nil, err
	}

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	container := &Container{
		Force:           force,
		Name:            name,
		Status:          status,
		Image:           image,
		Entrypoint:      renderedEntrypoint,
		Command:         renderedCommand,
		WorkingDir:      workDir,
		Env:             renderedEnv,
		Expose:          renderedExpose,
		Links:           renderedLinks,
		PublishAllPorts: publishAllPorts,
		PortBindings:    renderedPorts,
		DNS:             renderedDNS,
		Volumes:         renderedVolumes,
		VolumesFrom:     renderedVolumesFrom,
	}
	container.SetClient(dockerClient)
	return container, validateContainer(container)
}

func validateContainer(container *Container) error {
	if container.Status != "" {
		if !strings.EqualFold(container.Status, containerStatusRunning) &&
			!strings.EqualFold(container.Status, containerStatusCreated) {
			return errors.New("status must be 'running' or 'created'")
		}
	}
	return nil
}
