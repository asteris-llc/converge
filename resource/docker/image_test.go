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

package docker_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestImageInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(docker.Image))
}

func TestImageRepoTag(t *testing.T) {
	t.Parallel()

	type repoTagTest struct {
		*docker.Image
		Expected string
	}

	tests := []repoTagTest{
		{&docker.Image{Name: "ubuntu", Tag: "precise"}, "ubuntu:precise"},
		{&docker.Image{Name: "ubuntu"}, "ubuntu:latest"},
		{&docker.Image{Name: "gliderlabs/alpine", Tag: "3.3"}, "gliderlabs/alpine:3.3"},
	}

	for _, test := range tests {
		assert.Equal(t, test.Expected, test.Image.RepoTag())
	}
}

func TestImageCheckImageNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindImageFunc: func(string) (*dc.APIImages, error) {
			return nil, nil
		},
	}
	image := &docker.Image{Name: "ubuntu", Tag: "precise"}
	image.SetClient(c)

	status, err := image.Check()
	assert.Nil(t, err)
	assert.True(t, status.Changes())
	assert.Equal(t, "<image-missing>", status.Diffs()["image"].Original())
	assert.Equal(t, "ubuntu:precise", status.Diffs()["image"].Current())
	assert.Equal(t, "ubuntu:precise", status.Value())
}

func TestImageCheckImageNoChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindImageFunc: func(string) (*dc.APIImages, error) {
			return &dc.APIImages{}, nil
		},
	}
	image := &docker.Image{Name: "ubuntu", Tag: "precise"}
	image.SetClient(c)

	status, err := image.Check()
	assert.Nil(t, err)
	assert.False(t, status.Changes())
	assert.Equal(t, "ubuntu:precise", status.Diffs()["image"].Original())
	assert.Equal(t, "ubuntu:precise", status.Diffs()["image"].Current())
	assert.Equal(t, "ubuntu:precise", status.Value())
}

func TestImageCheckFailed(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindImageFunc: func(string) (*dc.APIImages, error) {
			return nil, errors.New("find image failed")
		},
	}
	image := &docker.Image{Name: "ubuntu", Tag: "precise"}
	image.SetClient(c)

	status, err := image.Check()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "find image failed")
	}
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.False(t, status.Changes())
}

func TestImageApply(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		PullImageFunc: func(string, string) error {
			return nil
		},
	}
	image := &docker.Image{Name: "ubuntu", Tag: "precise"}
	image.SetClient(c)

	assert.NoError(t, image.Apply())
}

type fakeAPIClient struct {
	FindImageFunc func(repoTag string) (*dc.APIImages, error)
	PullImageFunc func(name, tag string) error
}

func (f *fakeAPIClient) FindImage(repoTag string) (*dc.APIImages, error) {
	return f.FindImageFunc(repoTag)
}

func (f *fakeAPIClient) PullImage(name, tag string) error {
	return f.PullImageFunc(name, tag)
}
