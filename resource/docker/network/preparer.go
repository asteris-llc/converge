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

// +build !solaris

package network

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	dc "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

// Preparer for docker networks
//
// Network is responsible for managing Docker networks. It assumes that there is
// already a Docker daemon running on the system.
type Preparer struct {
	// name of the network
	Name string `hcl:"name" required:"true" nonempty:"true"`

	// network driver. default: bridge
	Driver string `hcl:"driver"`

	// labels to set on the network
	Labels map[string]string `hcl:"labels"`

	// driver specific options
	Options map[string]interface{} `hcl:"options"`

	// ip address management driver. default: default
	IPAMDriver string `hcl:"ipam_driver"`

	// optional custom IPAM configuration. multiple IPAM configurations are
	// permitted. Each IPAM configuration block should contain one or more of the
	// following items:
	//
	//   * subnet:      subnet in CIDR format
	//   * gateway:     ipv4 or ipv6 gateway for the corresponding subnet
	//   * ip_range:    container ips are allocated from this sub-ranges (CIDR format)
	//   * aux_address: auxiliary ipv4 or ipv6 addresses used by the network driver.
	//                  Aux addresses are specified as a map with a name key and an IP
	//                  address value
	IPAMConfig []ipamConfigMap `hcl:"ipam_config"`

	// restricts external access to the network
	Internal bool `hcl:"internal"`

	// enable ipv6 networking
	IPv6 bool `hcl:"ipv6"`

	// indicates whether the network should exist. default: present
	State State `hcl:"state" valid_values:"present,absent"`

	// indicates whether or not the network will be recreated if the state is not
	// what is expected. By default, the module will only check to see if the
	// network exists. Specified as a boolean value
	Force bool `hcl:"force"`
}

// Prepare a docker network
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if p.Driver == "" {
		p.Driver = DefaultDriver
	}

	if p.IPAMDriver == "" {
		p.IPAMDriver = DefaultIPAMDriver
	}

	if p.State == "" {
		p.State = StatePresent
	}

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	nw := &Network{
		Name:     p.Name,
		Driver:   p.Driver,
		Labels:   p.Labels,
		Options:  p.Options,
		IPAM:     p.buildIPAMOptions(),
		Internal: p.Internal,
		IPv6:     p.IPv6,
		State:    p.State,
		Force:    p.Force,
	}
	nw.SetClient(dockerClient)
	return nw, nil
}

func (p *Preparer) buildIPAMOptions() dc.IPAMOptions {
	ipamOptions := dc.IPAMOptions{
		Driver: p.IPAMDriver,
	}

	for _, ipamConfigMap := range p.IPAMConfig {
		ipamConfig := ipamConfigMap.IPAMConfig()
		if ipamConfig.Subnet != "" || len(ipamConfig.AuxAddress) > 0 {
			ipamOptions.Config = append(ipamOptions.Config, ipamConfig)
		}
	}

	return ipamOptions
}

type ipamConfigMap map[string]interface{}

func (i ipamConfigMap) IPAMConfig() dc.IPAMConfig {
	config := dc.IPAMConfig{}
	subnet := i.value("subnet")
	if subnet != "" {
		config.Subnet = subnet
		config.Gateway = i.value("gateway")
		config.IPRange = i.value("ip_range")
	}

	if val, ok := i["aux_addresses"]; ok {
		if auxMap, ok := val.(map[string]interface{}); ok {
			auxAddrs := make(map[string]string)
			for name, ipval := range auxMap {
				if ip, ok := ipval.(string); ok {
					auxAddrs[name] = ip
				}
			}
			config.AuxAddress = auxAddrs
		}
	}

	return config
}

func (i ipamConfigMap) value(key string) string {
	if val, ok := i[key]; ok {
		if strval, ok := val.(string); ok {
			return strval
		}
	}
	return ""
}

func init() {
	registry.Register("docker.network", (*Preparer)(nil), (*Network)(nil))
}
