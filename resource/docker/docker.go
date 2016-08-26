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
	"log"
	"strings"
	"time"

	dc "github.com/fsouza/go-dockerclient"
)

// APIClient provides access to docker
type APIClient interface {
	FindImage(string) (*dc.APIImages, error)
	PullImage(string, string) error
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
		return nil, err
	}
	return &Client{Client: c}, nil
}

// FindImage finds a local docker image with the specified repo tag
func (c *Client) FindImage(repoTag string) (*dc.APIImages, error) {
	images, err := c.Client.ListImages(dc.ListImagesOptions{Filter: repoTag})
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] docker: image filter %s found %d images", repoTag, len(images))
	for _, image := range images {
		for _, tag := range image.RepoTags {
			log.Printf("[DEBUG] docker: found %s", tag)
			if strings.EqualFold(repoTag, tag) {
				return &image, nil
			}
		}
	}

	return nil, nil
}

// PullImage pulls an image with the specified name and tag
func (c *Client) PullImage(name, tag string) error {
	log.Printf("[DEBUG] docker: pulling %s:%s", name, tag)
	opts := dc.PullImageOptions{
		Repository:        name,
		Tag:               tag,
		InactivityTimeout: c.PullInactivityTimeout,
	}

	err := c.Client.PullImage(opts, dc.AuthConfiguration{})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] docker: done pulling %s:%s", name, tag)
	return nil
}
