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
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

// APIClient provides access to docker
type APIClient interface {
	FindImage(string) (*dc.Image, error)
	PullImage(string, string) error
	FindContainer(string) (*dc.Container, error)
	CreateContainer(dc.CreateContainerOptions) (*dc.Container, error)
	StartContainer(string, string) error
}

// VolumeClient manages Docker volumes
type VolumeClient interface {
	FindVolume(string) (*dc.Volume, error)
	CreateVolume(dc.CreateVolumeOptions) (*dc.Volume, error)
	RemoveVolume(string) error
}

// Client provides api access to Docker
type Client struct {
	*dc.Client
	PullInactivityTimeout time.Duration
}

// NewDockerClient returns a docker client with the default configuration
func NewDockerClient() (*Client, error) {
	c, err := dc.NewClientFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create docker client from environment")
	}
	return &Client{Client: c}, nil
}

// FindImage finds a local docker image with the specified repo tag
func (c *Client) FindImage(repoTag string) (*dc.Image, error) {
	// TODO: can I just call inspect with the repoTag?
	images, err := c.Client.ListImages(dc.ListImagesOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find image")
	}

	log.WithFields(log.Fields{
		"module":      "docker",
		"filter":      repoTag,
		"image_count": len(images),
	}).Debug("image filter found images")

	var imageID string
	for _, image := range images {
		if repoTag == image.ID {
			imageID = image.ID
			break
		}

		for _, tag := range image.RepoTags {
			log.WithField("module", "docker").WithField("tag", tag).Debug("found tag")
			if strings.EqualFold(repoTag, tag) {
				imageID = image.ID
				break
			}
		}
		if imageID != "" {
			break
		}
	}

	if imageID != "" {
		log.WithField("module", "docker").WithField("tag", repoTag).Debug("found image")
		image, err := c.Client.InspectImage(imageID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to inspect image %s (%s)", repoTag, imageID)
		}

		return image, nil
	}

	log.WithField("module", "docker").WithField("tag", repoTag).Debug("could not find image")
	return nil, nil
}

// PullImage pulls an image with the specified name and tag
func (c *Client) PullImage(name, tag string) error {
	log.WithFields(log.Fields{
		"module": "docker",
		"name":   name,
		"tag":    tag,
	}).Debug("pulling")
	opts := dc.PullImageOptions{
		Repository:        name,
		Tag:               tag,
		InactivityTimeout: c.PullInactivityTimeout,
	}

	err := c.Client.PullImage(opts, dc.AuthConfiguration{})
	if err != nil {
		return errors.Wrap(err, "failed to pull image")
	}

	log.WithFields(log.Fields{
		"module": "docker",
		"name":   name,
		"tag":    tag,
	}).Debug("done pulling")
	return nil
}

// FindContainer returns a container matching the specified name
func (c *Client) FindContainer(name string) (*dc.Container, error) {
	opts := dc.ListContainersOptions{All: true}
	containers, err := c.Client.ListContainers(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list containers")
	}

	var containerID string
	for _, container := range containers {
		// check if ID was specified first
		if name == container.ID {
			containerID = container.ID
			break
		}

		// check container names for a match
		for _, cname := range container.Names {
			if strings.EqualFold(name, strings.TrimPrefix(cname, "/")) {
				containerID = container.ID
				break
			}
		}

		if containerID != "" {
			break
		}
	}

	if containerID != "" {
		log.WithField("module", "docker").WithField("name", name).Debug("found container")
		container, err := c.Client.InspectContainer(containerID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to inspect container %s (%s)", name, containerID)
		}

		return container, nil
	}

	log.WithField("module", "docker").WithField("name", name).Debug("could not find container")
	return nil, nil
}

// CreateContainer creates a container with the specified options
func (c *Client) CreateContainer(opts dc.CreateContainerOptions) (*dc.Container, error) {
	name := opts.Name

	container, err := c.FindContainer(name)
	if err != nil {
		return nil, err
	}

	// the container already exists
	if container != nil {
		log.WithField("module", "docker").WithField("name", name).Debug("container exists")

		// stop the container if running
		if container.State.Running {
			log.WithField("module", "docker").WithFields(log.Fields{"name": name, "id": container.ID}).Debug("stopping container")
			err = c.Client.StopContainer(container.ID, 60)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to stop container %s (%s)", name, container.ID)
			}
		}

		// remove the container
		log.WithField("module", "docker").WithFields(log.Fields{"name": name, "id": container.ID}).Debug("removing container")
		err = c.Client.RemoveContainer(dc.RemoveContainerOptions{ID: container.ID})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to remove container %s (%s)", name, container.ID)
		}
	}

	// create the container
	log.WithField("module", "docker").WithField("name", name).Debug("creating container")
	container, err = c.Client.CreateContainer(opts)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create container %s", name)
	}

	return container, err
}

// StartContainer starts the container with the specified ID
func (c *Client) StartContainer(name, containerID string) error {
	log.WithField("module", "docker").WithFields(log.Fields{"name": name, "id": containerID}).Debug("starting container")
	err := c.Client.StartContainer(containerID, nil)
	if err != nil {
		err = errors.Wrapf(err, "failed to start container %s (%s)", name, containerID)
	}
	return err
}

// CreateVolume creates a docker volume
func (c *Client) CreateVolume(opts dc.CreateVolumeOptions) (*dc.Volume, error) {
	log.WithFields(log.Fields{
		"module":   "docker",
		"function": "CreateVolume",
	}).Debugf("creating volume %s: %+v", opts.Name, opts)

	vol, err := c.Client.CreateVolume(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create volume")
	}
	return vol, nil
}

// RemoveVolume removes a docker volume
func (c *Client) RemoveVolume(name string) error {
	log.WithFields(log.Fields{
		"module":   "docker",
		"function": "RemoveVolume",
	}).Debugf("removing volume %s", name)

	err := c.Client.RemoveVolume(name)
	if err != nil {
		return errors.Wrap(err, "failed to remove volume")
	}
	return nil
}

// FindVolume finds the volume with the specified name
func (c *Client) FindVolume(name string) (*dc.Volume, error) {
	volume, err := c.Client.InspectVolume(name)
	if err != nil && err != dc.ErrNoSuchVolume {
		return nil, errors.Wrap(err, "failed to inspect volume")
	}
	return volume, nil
}
