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

package network

import (
	"fmt"
	"sort"
	"strings"

	"github.com/asteris-llc/converge/helpers/transform"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	dc "github.com/fsouza/go-dockerclient"
)

// State type for Network
type State string

const (
	// StatePresent indicates the network should be present
	StatePresent State = "present"

	// StateAbsent indicates the network should be absent
	StateAbsent State = "absent"

	// DefaultDriver is the default network driver
	DefaultDriver = "bridge"
)

// Network is responsible for managing docker volumes
type Network struct {
	*resource.Status
	client docker.NetworkClient

	Name    string
	Driver  string
	Labels  map[string]string
	Options map[string]interface{}
	State   State
	Force   bool
}

// Check system for docker network
func (n *Network) Check(resource.Renderer) (resource.TaskStatus, error) {
	n.Status = resource.NewStatus()
	nw, err := n.client.FindNetwork(n.Name)

	if err != nil {
		n.RaiseLevel(resource.StatusFatal)
		return n, err
	}

	n.AddDifference(n.Name, string(networkState(nw)), string(n.State), "")

	if n.State == StatePresent && nw != nil && n.Force {
		n.AddDifference("labels", mapCompareStr(nw.Labels), mapCompareStr(n.Labels), "")
		n.AddDifference("driver", nw.Driver, n.Driver, DefaultDriver)
	}

	if resource.AnyChanges(n.Differences) {
		n.RaiseLevel(resource.StatusWillChange)
	}

	return n, nil
}

// Apply ensures the network matches the desired state
func (n *Network) Apply() (resource.TaskStatus, error) {
	n.Status = resource.NewStatus()

	var (
		nw  *dc.Network
		err error
	)

	nw, err = n.client.FindNetwork(n.Name)
	if err != nil {
		n.RaiseLevel(resource.StatusFatal)
		return n, err
	}

	if n.State == StatePresent {
		if nw != nil {
			if !n.Force {
				return n, nil
			}

			err = n.client.RemoveNetwork(n.Name)
			if err != nil {
				n.RaiseLevel(resource.StatusFatal)
				return n, err
			}
			n.AddMessage(fmt.Sprintf("removed network %s", n.Name))
		}

		opts := dc.CreateNetworkOptions{
			Name:    n.Name,
			Driver:  n.Driver,
			Labels:  n.Labels,
			Options: n.Options,
		}
		nw, err = n.client.CreateNetwork(opts)
		if err != nil {
			n.RaiseLevel(resource.StatusFatal)
			return n, err
		}
		n.AddMessage(fmt.Sprintf("created network %s", n.Name))
		n.RaiseLevel(resource.StatusWillChange)
	} else {
		if nw != nil {
			err = n.client.RemoveNetwork(n.Name)
			if err != nil {
				n.RaiseLevel(resource.StatusFatal)
				return n, err
			}
			n.AddMessage(fmt.Sprintf("removed network %s", n.Name))
			n.RaiseLevel(resource.StatusWillChange)
		}
	}

	n.AddDifference(n.Name, string(networkState(nw)), string(n.State), "")
	return n, nil
}

// SetClient injects a docker api client
func (n *Network) SetClient(client docker.NetworkClient) {
	n.client = client
}

func networkState(nw *dc.Network) State {
	if nw != nil {
		return StatePresent
	}
	return StateAbsent
}

func mapCompareStr(m map[string]string) string {
	pairs := transform.StringsMapToStringSlice(
		m,
		func(k, v string) string {
			return fmt.Sprintf("%s=%s", k, v)
		},
	)
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}
