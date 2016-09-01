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

	// allocates a random host port for all of a container’s exposed ports. Specified as a boolean value
	PublishAllPorts bool `hcl:"publish_all_ports"` // TODO: how do we render bool values from params

	// the desired status of the container. running|created
	Status string

	// indicates whether or not the container will be recreated if the state is
	// not what is expected. By default, the module will only check to see if the
	// container exists
	Force bool
}

// Prepare a docker container
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	name, err := requiredRender(render, "name", p.Name)
	if err != nil {
		return nil, err
	}

	image, err := requiredRender(render, "image", p.Image)
	if err != nil {
		return nil, err
	}

	renderedEntrypoint, err := renderStringSlice(render, "entrypoint", p.Entrypoint)
	if err != nil {
		return nil, err
	}

	renderedCommand, err := renderStringSlice(render, "command", p.Command)
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

	renderedEnv, err := renderStringMapToStringSlice(render, "env", p.Env, func(k, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	})
	if err != nil {
		return nil, err
	}

	renderedExpose, err := renderStringSlice(render, "expose", p.Expose)
	if err != nil {
		return nil, err
	}

	renderedLinks, err := renderStringSlice(render, "links", p.Links)
	if err != nil {
		return nil, err
	}

	renderedPorts, err := renderStringSlice(render, "ports", p.Ports)
	if err != nil {
		return nil, err
	}

	renderedDNS, err := renderStringSlice(render, "dns", p.DNS)
	if err != nil {
		return nil, err
	}

	renderedVolumes, err := renderStringSlice(render, "volumes", p.Volumes)
	if err != nil {
		return nil, err
	}

	renderedVolumesFrom, err := renderStringSlice(render, "volumes_from", p.VolumesFrom)
	if err != nil {
		return nil, err
	}

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	container := &Container{
		Force:           p.Force,
		Name:            name,
		Status:          status,
		Image:           image,
		Entrypoint:      renderedEntrypoint,
		Command:         renderedCommand,
		WorkingDir:      workDir,
		Env:             renderedEnv,
		Expose:          renderedExpose,
		Links:           renderedLinks,
		PublishAllPorts: p.PublishAllPorts,
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

func requiredRender(render resource.Renderer, name, content string) (string, error) {
	rendered, err := render.Render(name, content)
	if err != nil {
		return "", err
	}

	if rendered == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return rendered, nil
}

func renderStringSlice(render resource.Renderer, name string, content []string) ([]string, error) {
	renderedSlice := make([]string, len(content))
	for i, val := range content {
		rendered, err := render.Render(fmt.Sprintf("%s[%d]", name, i), val)
		if err != nil {
			return nil, err
		}
		renderedSlice[i] = rendered
	}
	return renderedSlice, nil
}

func renderStringMapToStringSlice(render resource.Renderer, name string, content map[string]string, stringFunc func(string, string) string) ([]string, error) {
	renderedSlice := make([]string, len(content))
	idx := 0
	for key, val := range content {
		pair := stringFunc(key, val)
		rendered, err := render.Render(fmt.Sprintf("%s[%s]", name, val), pair)
		if err != nil {
			return nil, err
		}
		renderedSlice[idx] = rendered
		idx++
	}

	return renderedSlice, nil
}
