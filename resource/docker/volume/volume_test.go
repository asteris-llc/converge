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

package volume_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker/volume"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestVolumeInterface verifies that Volume implements the resource.Task
// interface
func TestVolumeInterface(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	assert.Implements(t, (*resource.Task)(nil), new(volume.Volume))
}

// TestVolumeCheck tests the Volume.Check function
func TestVolumeCheck(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("state: absent", func(t *testing.T) {
		t.Run("volume does not exist", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "absent"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(nil, nil)

			status, err := vol.Check(context.Background(), fakerenderer.New())
			assert.Nil(t, err)
			assert.False(t, status.HasChanges())
			assert.Equal(t, 1, len(status.Diffs()))
			comparison.AssertDiff(t, status.Diffs(), "test-volume", "absent", "absent")
		})

		t.Run("volume exists", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "absent"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)

			status, err := vol.Check(context.Background(), fakerenderer.New())
			assert.Nil(t, err)
			assert.True(t, status.HasChanges())
			assert.Equal(t, 1, len(status.Diffs()))
			comparison.AssertDiff(t, status.Diffs(), "test-volume", "present", "absent")
		})
	})

	t.Run("state: present", func(t *testing.T) {
		t.Run("volume does not exist", func(t *testing.T) {
			vol := &volume.Volume{
				Name:  "test-volume",
				State: "present",
			}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(nil, nil)

			status, err := vol.Check(context.Background(), fakerenderer.New())
			assert.Nil(t, err)
			assert.True(t, status.HasChanges())
			assert.Equal(t, 1, len(status.Diffs()))
			comparison.AssertDiff(t, status.Diffs(), "test-volume", "absent", "present")
		})

		t.Run("volume exists", func(t *testing.T) {
			vol := &volume.Volume{
				Name:  "test-volume",
				State: "present",
			}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)

			status, err := vol.Check(context.Background(), fakerenderer.New())
			assert.Nil(t, err)
			assert.False(t, status.HasChanges())
			assert.Equal(t, 1, len(status.Diffs()))
			comparison.AssertDiff(t, status.Diffs(), "test-volume", "present", "present")
		})

		t.Run("volume exists, force: true", func(t *testing.T) {
			t.Run("labels", func(t *testing.T) {
				vol := &volume.Volume{
					Name:   "test-volume",
					State:  "present",
					Labels: map[string]string{"key": "val", "test": "val2"},
					Force:  true,
				}
				c := &mockClient{}
				vol.SetClient(c)
				c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)

				status, err := vol.Check(context.Background(), fakerenderer.New())
				assert.Nil(t, err)
				assert.True(t, status.HasChanges())
				assert.True(t, len(status.Diffs()) > 1)
				comparison.AssertDiff(t, status.Diffs(), "test-volume", "present", "present")
				comparison.AssertDiff(t, status.Diffs(), "labels", "", "key=val, test=val2")
			})

			t.Run("driver", func(t *testing.T) {
				vol := &volume.Volume{
					Name:   "test-volume",
					State:  "present",
					Driver: "flocker",
					Force:  true,
				}
				c := &mockClient{}
				vol.SetClient(c)
				c.On("FindVolume", "test-volume").
					Return(&dc.Volume{Name: "test-volume", Driver: "local"}, nil)
				vol.SetClient(c)

				status, err := vol.Check(context.Background(), fakerenderer.New())
				assert.Nil(t, err)
				assert.True(t, status.HasChanges())
				assert.True(t, len(status.Diffs()) > 1)
				comparison.AssertDiff(t, status.Diffs(), "test-volume", "present", "present")
				comparison.AssertDiff(t, status.Diffs(), "driver", "local", "flocker")
			})
		})
	})

	t.Run("docker api error", func(t *testing.T) {
		vol := &volume.Volume{Name: "test-volume", State: "present"}
		c := &mockClient{}
		vol.SetClient(c)
		c.On("FindVolume", "test-volume").Return(nil, errors.New("error"))

		status, err := vol.Check(context.Background(), fakerenderer.New())
		require.Error(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
	})
}

// TestVolumeApply tests the Volume.Apply function
func TestVolumeApply(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("docker find volume error", func(t *testing.T) {
		vol := &volume.Volume{Name: "test-volume"}
		c := &mockClient{}
		vol.SetClient(c)
		c.On("FindVolume", "test-volume").Return(nil, errors.New("error"))

		status, err := vol.Apply(context.Background())
		require.Error(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
	})

	t.Run("state: absent", func(t *testing.T) {
		t.Run("volume exists", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "absent"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)
			c.On("RemoveVolume", mock.AnythingOfType("string")).Return(nil)

			status, err := vol.Apply(context.Background())
			require.NoError(t, err)
			assert.True(t, status.HasChanges())
			c.AssertCalled(t, "RemoveVolume", "test-volume")
		})

		t.Run("volume does not exist", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "absent"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(nil, nil)

			status, err := vol.Apply(context.Background())
			require.NoError(t, err)
			assert.False(t, status.HasChanges())
			c.AssertNotCalled(t, "RemoveVolume", "test-volume")
		})

		t.Run("docker remove volume error", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "absent"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)
			c.On("RemoveVolume", mock.AnythingOfType("string")).Return(errors.New("test-volume"))

			status, err := vol.Apply(context.Background())
			require.Error(t, err)
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
		})
	})

	t.Run("state: present", func(t *testing.T) {
		t.Run("volume does not exist", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "present"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(nil, nil)
			c.On("CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions")).
				Return(&dc.Volume{Name: "test-volume"}, nil)

			status, err := vol.Apply(context.Background())
			require.NoError(t, err)
			assert.True(t, status.HasChanges())
			c.AssertCalled(t, "CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions"))
			c.AssertNotCalled(t, "RemoveVolume", "test-volume")
		})

		t.Run("volume exists, force: false", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "present"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(&dc.Volume{Name: "test-volume"}, nil)

			status, err := vol.Apply(context.Background())
			require.NoError(t, err)
			assert.False(t, status.HasChanges())
			c.AssertNotCalled(t, "CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions"))
			c.AssertNotCalled(t, "RemoveVolume", "test-volume")
		})

		t.Run("volume exists, force: true", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "present", Force: true}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").
				Return(&dc.Volume{Name: "test-volume", Driver: "flocker"}, nil)
			c.On("RemoveVolume", mock.AnythingOfType("string")).Return(nil)
			c.On("CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions")).
				Return(&dc.Volume{Name: "test-volume"}, nil)

			status, err := vol.Apply(context.Background())
			require.NoError(t, err)
			assert.True(t, status.HasChanges())
			c.AssertCalled(t, "RemoveVolume", "test-volume")
			c.AssertCalled(t, "CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions"))
		})

		t.Run("docker create volume error", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "present"}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").Return(nil, nil)
			c.On("CreateVolume", mock.AnythingOfType("docker.CreateVolumeOptions")).
				Return(nil, errors.New("error"))

			status, err := vol.Apply(context.Background())
			require.Error(t, err)
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
		})

		t.Run("docker remove volume error", func(t *testing.T) {
			vol := &volume.Volume{Name: "test-volume", State: "present", Force: true}
			c := &mockClient{}
			vol.SetClient(c)
			c.On("FindVolume", "test-volume").
				Return(&dc.Volume{Name: "test-volume", Driver: "flocker"}, nil)
			c.On("RemoveVolume", mock.AnythingOfType("string")).Return(errors.New("error"))

			status, err := vol.Apply(context.Background())
			require.Error(t, err)
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
		})
	})
}

type mockClient struct {
	mock.Mock
}

func (m *mockClient) FindVolume(name string) (*dc.Volume, error) {
	args := m.Called(name)

	ret := args.Get(0)
	if ret == nil {
		return nil, args.Error(1)
	}

	return ret.(*dc.Volume), args.Error(1)
}

func (m *mockClient) CreateVolume(opts dc.CreateVolumeOptions) (*dc.Volume, error) {
	args := m.Called(opts)
	ret := args.Get(0)
	if ret == nil {
		return nil, args.Error(1)
	}
	return ret.(*dc.Volume), args.Error(1)
}

func (m *mockClient) RemoveVolume(name string) error {
	args := m.Called(name)
	return args.Error(0)
}
