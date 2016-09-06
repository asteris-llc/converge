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

package container_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker/container"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestContainerInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(container.Container))
}

func TestContainerCheckContainerNotFound(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(string) (*dc.Container, error) {
			return nil, nil
		},
	}

	name := "nginx"
	container := &container.Container{Force: true, Name: name}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assert.Equal(t, name, status.Value())
	assertDiff(t, status.Diffs(), "name", "<container-missing>", name)
}

func TestContainerCheckContainerFindContainerError(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(string) (*dc.Container, error) {
			return nil, errors.New("find container failed")
		},
	}

	container := &container.Container{Force: true, Name: "nginx"}
	container.SetClient(c)

	status, err := container.Check()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "find container failed")
	}
	assert.Equal(t, resource.StatusWillChange, status.StatusCode())
	assert.True(t, status.HasChanges())
}

func TestContainerCheckContainerNoChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(string) (*dc.Container, error) {
			return &dc.Container{
				Name:   "nginx",
				State:  dc.State{Status: "running"},
				Config: &dc.Config{}}, nil
		},
		FindImageFunc: func(string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx"}
	container.SetClient(c)

	status, err := container.Check()

	assert.Nil(t, err)
	assert.False(t, status.HasChanges())
}

func TestContainerCheckStatusNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		// the existing container is running the "nginx" command
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Cmd: []string{"nginx"},
				},
				State: dc.State{Status: "exited"},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx"}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "status", "exited", "running")
}

func TestContainerCheckStatusNoChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name:   name,
				Config: &dc.Config{},
				State:  dc.State{Status: "created"},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx", Status: "created"}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.False(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "status", "created", "created")
}

func TestContainerCheckCommandNeedsChange(t *testing.T) {
	// This test simulates a running container with a command that is different
	// than the specified command
	t.Parallel()

	c := &fakeAPIClient{
		// the existing container is running the "nginx" command
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Cmd: []string{"nginx"},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{
		Force:   true,
		Name:    "nginx",
		Command: []string{"nginx", "-g", "daemon", "off;"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "command", "nginx", "nginx -g daemon off;")
}

func TestContainerCheckEmptyCommandNeedsChange(t *testing.T) {
	// Specifying an empty Command means we want to use the default image command.
	// This test simulates a running container with a command that is different
	// than the image default.
	t.Parallel()

	c := &fakeAPIClient{
		// the existing container is running the "nginx" command
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Cmd: []string{"nginx"},
				},
			}, nil
		},
		// the image has a default command of "nginx -g daemon off;"
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{
					Cmd: []string{"nginx", "-g", "daemon off;"},
				},
			}, nil
		},
	}

	// the resource uses an empty command implying that the default should be
	// running
	container := &container.Container{
		Force:   true,
		Name:    "nginx",
		Command: []string{},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "command", "nginx", "nginx -g daemon off;")
}

func TestContainerCheckImageNeedsChange(t *testing.T) {
	// This test simulates a running container with an image that is different
	// than the specified image.
	t.Parallel()

	c := &fakeAPIClient{
		// the existing container is running the "nginx" image
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name:   name,
				Image:  "nginx",
				Config: &dc.Config{},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				RepoTags: []string{"nginx"},
				Config: &dc.Config{
					Image: "nginx",
				},
			}, nil
		},
	}

	// the resource uses a different image
	container := &container.Container{Force: true, Name: "nginx", Image: "busybox"}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "image", "nginx", "busybox")
}

func TestContainerCheckEntrypointNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		// the existing container defaults to the "start" entrypoint
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Entrypoint: []string{},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{
				Entrypoint: []string{"start"},
			}}, nil
		},
	}

	container := &container.Container{
		Force:      true,
		Name:       "nginx",
		Entrypoint: []string{"/bin/bash", "start"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "entrypoint", "start", "/bin/bash start")
}

func TestContainerCheckWorkingDirNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					WorkingDir: "/tmp",
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx", WorkingDir: "/tmp/working"}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "working_dir", "/tmp", "/tmp/working")
}

func TestContainerCheckEnvNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Env: []string{
						"PATH=/usr/bin",                    // set from image
						"HTTP_PROXY=http://localhost:8080", // set by engine
						"no_proxy=*.local",                 // set by engine
						"FROMIMAGE=yes",                    // set from image
						"FOO=BAR",                          // set in container
						"EXTRA=TEST",                       // set in container
					},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{
					Env: []string{"PATH=/usr/bin", "FROMIMAGE=yes"},
				},
			}, nil
		},
	}

	container := &container.Container{
		Force: true,
		Name:  "nginx",
		Env: []string{
			"BAR=BAZ",                      // new container var
			"FOO=BAR",                      // existing container var
			"PATH=/usr/bin;/usr/sbin",      // override image var
			"NO_PROXY=*.local, 169.254/16", // override engine var
		},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	// diff should include the new BAR var and the overridden PATH and NO_PROXY
	// vars. The EXTRA var should not be included in the desired state either
	assertDiff(
		t,
		status.Diffs(),
		"env",
		"EXTRA=TEST FOO=BAR PATH=/usr/bin no_proxy=*.local",
		"BAR=BAZ FOO=BAR NO_PROXY=*.local, 169.254/16 PATH=/usr/bin;/usr/sbin",
	)
}

func TestContainerCheckExposeNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					ExposedPorts: map[dc.Port]struct{}{
						"80/tcp":  struct{}{},
						"443/tcp": struct{}{},
					},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{
					ExposedPorts: map[dc.Port]struct{}{
						"80/tcp":  struct{}{},
						"443/tcp": struct{}{},
					},
				},
			}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx", Expose: []string{"8001", "8002/udp"}}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "expose", "443/tcp, 80/tcp", "443/tcp, 80/tcp, 8001/tcp, 8002/udp")
}

func TestContainerCheckPortsNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					ExposedPorts: map[dc.Port]struct{}{
						"80/tcp":   struct{}{},
						"443/tcp":  struct{}{},
						"8003/tcp": struct{}{},
					},
				},
				HostConfig: &dc.HostConfig{
					PortBindings: map[dc.Port][]dc.PortBinding{
						dc.Port("80/tcp"): []dc.PortBinding{dc.PortBinding{HostPort: "8003"}},
					},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{
					ExposedPorts: map[dc.Port]struct{}{
						"80/tcp":  struct{}{},
						"443/tcp": struct{}{},
					},
				},
			}, nil
		},
	}

	container := &container.Container{
		Force:        true,
		Name:         "nginx",
		Expose:       []string{"8003", "8005/udp"},
		PortBindings: []string{"127.0.0.1:8000:80", "127.0.0.1::80/tcp", "443:443", "8003:80", "8004:80", "80", "8085/udp"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "ports", ":8003:80/tcp", "127.0.0.1:8000:80/tcp, 127.0.0.1::80/tcp, :443:443/tcp, :8003:80/tcp, :8004:80/tcp, ::80/tcp, ::8085/udp")
}

func TestContainerCheckLinksNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name:   name,
				Config: &dc.Config{},
				HostConfig: &dc.HostConfig{
					// no alias
					Links: []string{fmt.Sprintf("/redis-server:/%s/redis-server", name)},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{},
			}, nil
		},
	}

	// include alias for existing link and a acouple of more links
	container := &container.Container{
		Force: true,
		Name:  "nginx",
		Links: []string{"redis-server:redis", "memcached", "postgresql:db"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "links",
		"redis-server",
		"memcached, postgresql:db, redis-server:redis")
}

func TestContainerCheckDNSNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name:   name,
				Config: &dc.Config{},
				HostConfig: &dc.HostConfig{
					DNS: []string{},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{
				Config: &dc.Config{},
			}, nil
		},
	}

	// include alias for existing link and a acouple of more links
	container := &container.Container{
		Force: true,
		Name:  "nginx",
		DNS:   []string{"8.8.8.8", "8.8.4.4"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "dns", "", "8.8.8.8, 8.8.4.4")
}

func TestContainerCheckVolumesNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Volumes: map[string]struct{}{
						"/var/log": struct{}{},
					},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{
				Volumes: map[string]struct{}{
					"/var/log": struct{}{},
				},
			}}, nil
		},
	}

	container := &container.Container{Force: true, Name: "nginx", Volumes: []string{"/var/html"}}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "volumes", "/var/log", "/var/html, /var/log")
}

func TestContainerCheckBindsNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name: name,
				Config: &dc.Config{
					Volumes: map[string]struct{}{
						"/var/log": struct{}{},
					},
				},
				HostConfig: &dc.HostConfig{
					Binds: []string{},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{
				Volumes: map[string]struct{}{
					"/var/log": struct{}{},
				},
			}}, nil
		},
	}

	container := &container.Container{
		Force:   true,
		Name:    "nginx",
		Volumes: []string{"/var/log:/var/log", "/var/db:/var/db:ro"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "volumes", "/var/log", "/var/db, /var/log")
	assertDiff(t, status.Diffs(), "binds", "", "/var/db:/var/db:ro, /var/log:/var/log")
}

func TestContainerCheckVolumesFromNeedsChange(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		FindContainerFunc: func(name string) (*dc.Container, error) {
			return &dc.Container{
				Name:   name,
				Config: &dc.Config{},
				HostConfig: &dc.HostConfig{
					VolumesFrom: []string{},
				},
			}, nil
		},
		FindImageFunc: func(repoTag string) (*dc.Image, error) {
			return &dc.Image{Config: &dc.Config{}}, nil
		},
	}

	container := &container.Container{
		Force:       true,
		Name:        "nginx",
		VolumesFrom: []string{"dbvol", "webvol:ro,z"},
	}
	container.SetClient(c)

	status, err := container.Check()
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assertDiff(t, status.Diffs(), "volumes_from", "", "dbvol, webvol:ro,z")
}

func TestContainerApply(t *testing.T) {
	t.Parallel()

	c := &fakeAPIClient{
		CreateContainerFunc: func(opts dc.CreateContainerOptions) (*dc.Container, error) {
			return &dc.Container{}, nil
		},
		StartContainerFunc: func(string, string) error { return nil },
	}
	container := &container.Container{Force: true, Name: "nginx", Image: "nginx:latest"}
	container.SetClient(c)

	assert.NoError(t, container.Apply())
}

func assertDiff(t *testing.T, diffs map[string]resource.Diff, name, original, current string) bool {
	var ok bool

	if ok = assert.NotEmpty(t, diffs); !ok {
		return false
	}

	if ok = assert.NotNil(t, diffs[name]); !ok {
		return false
	}

	if ok = assert.Equal(t, original, diffs[name].Original()); !ok {
		return false
	}

	if ok = assert.Equal(t, current, diffs[name].Current()); !ok {
		return false
	}

	return true
}

type fakeAPIClient struct {
	FindImageFunc       func(repoTag string) (*dc.Image, error)
	PullImageFunc       func(name, tag string) error
	FindContainerFunc   func(name string) (*dc.Container, error)
	CreateContainerFunc func(opts dc.CreateContainerOptions) (*dc.Container, error)
	StartContainerFunc  func(name, id string) error
}

func (f *fakeAPIClient) FindImage(repoTag string) (*dc.Image, error) {
	return f.FindImageFunc(repoTag)
}

func (f *fakeAPIClient) PullImage(name, tag string) error {
	return f.PullImageFunc(name, tag)
}

func (f *fakeAPIClient) FindContainer(name string) (*dc.Container, error) {
	return f.FindContainerFunc(name)
}

func (f *fakeAPIClient) CreateContainer(opts dc.CreateContainerOptions) (*dc.Container, error) {
	return f.CreateContainerFunc(opts)
}

func (f *fakeAPIClient) StartContainer(name, id string) error {
	return f.StartContainerFunc(name, id)
}

func (f *fakeAPIClient) SetRetryOptions(timeout, retryInterval time.Duration) {}
