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

	dc "github.com/fsouza/go-dockerclient"
)

// APIClient provides access to docker
type APIClient interface {
	FindImage(string) (*dc.APIImages, error)
	PullImage(string, string) error
}

type dockerClient struct {
	*dc.Client
}

func newDockerClient() (APIClient, error) {
	c, err := dc.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &dockerClient{c}, nil
}

func (c *dockerClient) FindImage(repoTag string) (*dc.APIImages, error) {
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

func (c *dockerClient) PullImage(name, tag string) error {
	log.Printf("[DEBUG] docker: pulling %s:%s", name, tag)
	opts := dc.PullImageOptions{
		Repository: name,
		Tag:        tag,
	}

	err := c.Client.PullImage(opts, dc.AuthConfiguration{})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] docker: done pulling %s:%s", name, tag)
	return nil
}
