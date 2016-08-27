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
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

// Container is responsible for running docker containers
type Container struct {
	Name       string
	Image      string
	Entrypoint string
	Command    string
	WorkingDir string
	client     docker.APIClient
}

// Check that a docker container with the specified configuration is running
func (c *Container) Check() (resource.TaskStatus, error) {
	status := &resource.Status{Status: c.Name}

	container, err := c.client.FindContainer(c.Name)
	if err != nil {
		status.WarningLevel = resource.StatusFatal
		return status, err
	}

	if container != nil {
		c.diffContainer(container, status)
	} else {
		status.AddDifference("name", "", c.Name, "<container-missing>")
	}

	if resource.AnyChanges(status.Differences) {
		status.WillChange = true
		status.WarningLevel = resource.StatusWillChange
	}

	return status, nil
}

// Apply starts a docker container with the specified configuration
func (c *Container) Apply() error {
	config := &dc.Config{
		Image:      c.Image,
		WorkingDir: c.WorkingDir,
	}

	if c.Command != "" {
		config.Cmd = strings.Split(c.Command, " ")
	}

	if c.Entrypoint != "" {
		config.Entrypoint = strings.Split(c.Entrypoint, " ")
	}

	opts := dc.CreateContainerOptions{
		Name:   c.Name,
		Config: config,
	}
	_, err := c.client.CreateContainer(opts)

	if err != nil {
		return errors.Wrapf(err, "failed to run container %s", c.Name)
	}
	return nil
}

// SetClient injects a docker api client
func (c *Container) SetClient(client docker.APIClient) {
	c.client = client
}

func (c *Container) diffContainer(container *dc.Container, status *resource.Status) error {
	status.AddDifference("name", strings.TrimPrefix(container.Name, "/"), c.Name, "")
	status.AddDifference("status", container.State.Status, "running", "")

	image, err := c.client.FindImage(container.Image)
	if err != nil {
		status.WarningLevel = resource.StatusFatal
		return errors.Wrapf(err, "failed to find image %s for container %s", container.Image, container.Name)
	}

	var actual, expected string
	// if Cmd is empty, compare using the default from the container Image
	actual = strings.Join(container.Config.Cmd, " ")
	if c.Command == "" {
		expected = strings.Join(image.Config.Cmd, " ")
	} else {
		expected = c.Command
	}
	status.AddDifference("command", actual, expected, "")

	// if Entrypoint is empty, compare using the default from the container Image
	actual = strings.Join(container.Config.Entrypoint, " ")
	if c.Entrypoint == "" {
		expected = strings.Join(image.Config.Entrypoint, " ")
	} else {
		expected = c.Entrypoint
	}
	status.AddDifference("entrypoint", actual, expected, "")

	// if WorkingDir is empty, compare using the default from the container Image
	actual = container.Config.WorkingDir
	if c.WorkingDir == "" {
		expected = image.Config.WorkingDir
	} else {
		expected = c.WorkingDir
	}
	status.AddDifference("working_dir", actual, expected, "")

	// look for a tag name instead of reporting the Image ID in the diff
	existingRepoTag := preferredRepoTag(c.Image, image)
	status.AddDifference("image", existingRepoTag, c.Image, "")

	return nil
}

func preferredRepoTag(want string, image *dc.Image) string {
	// look for the matching repo tag in case an image has multiple tags. if there
	// is no matching, return the first if it has one
	var existingRepoTag string
	if len(image.RepoTags) > 0 {
		for _, repoTag := range image.RepoTags {
			if strings.EqualFold(want, repoTag) {
				existingRepoTag = repoTag
				break
			}
		}

		if existingRepoTag == "" {
			existingRepoTag = image.RepoTags[0]
		}
	}

	return existingRepoTag
}
