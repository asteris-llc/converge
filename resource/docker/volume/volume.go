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

package volume

import (
	"fmt"
	"sort"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	dc "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

// State type for Volume
type State string

const (
	// StatePresent indicates the volume should be present
	StatePresent State = "present"

	// StateAbsent indicates the volume should be absent
	StateAbsent State = "absent"
)

// Volume is responsible for managing docker volumes
type Volume struct {
	*resource.Status
	client docker.VolumeClient

	Name    string
	Labels  map[string]string
	Driver  string
	Options map[string]string
	State   State
	Force   bool
}

// Check system for docker volume
func (v *Volume) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	v.Status = resource.NewStatus()
	vol, err := v.client.FindVolume(v.Name)

	if err != nil {
		v.Status.Level = resource.StatusFatal
		return v, err
	}

	v.Status.AddDifference(v.Name, string(volumeState(vol)), string(v.State), "")

	if v.State == StatePresent && vol != nil && v.Force {
		expectedLabels := mapToString(v.Labels)
		actualLabels := mapToString(vol.Labels)
		v.Status.AddDifference("labels", actualLabels, expectedLabels, "")
		v.Status.AddDifference("driver", vol.Driver, v.Driver, "local")
		// we cannot detect difference in Options because the Docker API does not
		// return that information:
		// https://docs.docker.com/engine/reference/api/docker_remote_api_v1.24/#/inspect-a-volume
	}

	if resource.AnyChanges(v.Status.Differences) {
		v.RaiseLevel(resource.StatusWillChange)
	}

	return v, nil
}

// Apply ensures the volume matches the desired state
func (v *Volume) Apply(context.Context) (resource.TaskStatus, error) {
	v.Status = resource.NewStatus()

	var (
		vol *dc.Volume
		err error
	)

	vol, err = v.client.FindVolume(v.Name)

	if err != nil {
		v.Status.Level = resource.StatusFatal
		return v, err
	}

	if v.State == StatePresent {
		if vol != nil {
			if !v.Force {
				return v, nil
			}

			err = v.client.RemoveVolume(v.Name)
			if err != nil {
				v.Status.Level = resource.StatusFatal
				return v, err
			}
			v.Status.AddMessage(fmt.Sprintf("removed volume %s", v.Name))
		}

		opts := dc.CreateVolumeOptions{
			Name:       v.Name,
			Driver:     v.Driver,
			DriverOpts: v.Options,
			Labels:     v.Labels,
		}

		vol, err = v.client.CreateVolume(opts)
		if err != nil {
			v.Status.Level = resource.StatusFatal
			return v, err
		}
		v.Status.AddMessage(fmt.Sprintf("created volume %s", v.Name))
		v.RaiseLevel(resource.StatusWillChange)
	} else {
		if vol != nil {
			err = v.client.RemoveVolume(v.Name)
			if err != nil {
				v.Status.Level = resource.StatusFatal
				return v, err
			}
			v.Status.AddMessage(fmt.Sprintf("removed volume %s", v.Name))
			v.RaiseLevel(resource.StatusWillChange)
		}
	}

	v.Status.AddDifference(v.Name, string(volumeState(vol)), string(v.State), "")
	return v, nil
}

// SetClient injects a docker api client
func (v *Volume) SetClient(client docker.VolumeClient) {
	v.client = client
}

func volumeState(vol *dc.Volume) State {
	if vol != nil {
		return StatePresent
	}
	return StateAbsent
}

func mapToString(m map[string]string) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	strs := make([]string, len(m))
	i := 0
	for k, v := range m {
		strs[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	sort.Strings(strs)
	return strings.Join(strs, ", ")
}
