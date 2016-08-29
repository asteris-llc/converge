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
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
)

// Preparer for docker containers
type Preparer struct {
	Name       string            `hcl:"name"`
	Image      string            `hcl:"image"`
	Entrypoint string            `hcl:"entrypoint"`
	Command    string            `hcl:"command"`
	WorkingDir string            `hcl:"working_dir"`
	Env        map[string]string `hcl:"env"`
	Expose     []string          `hcl:"expose"`

	// Allocates a random host port for all of a container’s exposed ports. Specified as a boolean value.
	PublishAllPorts bool `hcl:"publish_all_ports"` // TODO: how do we render bool values from params
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

	entrypoint, err := render.Render("entrypoint", p.Entrypoint)
	if err != nil {
		return nil, err
	}

	command, err := render.Render("command", p.Command)
	if err != nil {
		return nil, err
	}

	workDir, err := render.Render("working_dir", p.WorkingDir)
	if err != nil {
		return nil, err
	}

	// render Env
	renderedEnv := make([]string, len(p.Env))
	idx := 0
	for name, val := range p.Env {
		pair := fmt.Sprintf("%s=%s", name, val)
		rendered, rerr := render.Render(fmt.Sprintf("env[%s]", name), pair)
		if rerr != nil {
			return nil, rerr
		}
		renderedEnv[idx] = rendered
		idx++
	}

	renderedExpose := make([]string, len(p.Expose))
	for i, expose := range p.Expose {
		rendered, rerr := render.Render(fmt.Sprintf("expose[%d]", i), expose)
		if rerr != nil {
			return nil, rerr
		}
		renderedExpose[i] = rendered
	}

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	container := &Container{
		Name:            name,
		Image:           image,
		Entrypoint:      entrypoint,
		Command:         command,
		WorkingDir:      workDir,
		Env:             renderedEnv,
		Expose:          renderedExpose,
		PublishAllPorts: p.PublishAllPorts,
	}
	container.SetClient(dockerClient)
	return container, nil
}

func requiredRender(render resource.Renderer, name string, content string) (string, error) {
	rendered, err := render.Render(name, content)
	if err != nil {
		return "", err
	}

	if rendered == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return rendered, nil
}
