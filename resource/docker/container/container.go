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

package container

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker"
	mapset "github.com/deckarep/golang-set"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const (
	containerStatusRunning = "running"

	// DefaultNetworkMode is the mode of the container network
	DefaultNetworkMode = "default"
)

// these variable names can be injected by the docker engine
var (
	engineEnvVars   = []string{"https_proxy", "http_proxy", "no_proxy", "ftp_proxy"}
	builtinNetworks = []string{"default", "bridge", "host", "none", "container"}
)

// Container is responsible for creating docker containers
type Container struct {
	// the name of the container
	Name string `export:"name"`

	// the name of the image
	Image string `export:"image"`

	// the entrypoint into the container
	Entrypoint []string `export:"entrypoint"`

	// the command to run
	Command []string `export:"command"`

	// the working directory
	WorkingDir string `export:"workingdir"`

	// configured environment variables for the container
	Env []string `export:"env"`

	// additional ports to exposed in the container
	Expose []string `export:"expose"`

	// A list of links for the container in the form of container_name:alias
	Links []string `export:"links"`

	// ports to bind
	PortBindings []string `export:"portbindings"`

	// list of DNS servers the container is using
	DNS []string `export:"dns"`

	// volumes that have been bind-mounted
	Volumes []string `export:"volumes"`

	// containers from which volumes have been mounted
	VolumesFrom []string `export:"volumesfrom"`

	// if true, all ports have been published
	PublishAllPorts bool `export:"publishallports"`

	// the mode of the container network
	NetworkMode string `export:"networkmode"`

	// networks the container is connected to
	Networks []string `export:"networks"`

	// the status of the container.
	CStatus string `export:"status"`

	// Indicate whether the 'force' flag was set
	Force  bool `export:"force"`
	client docker.APIClient
}

// Check that a docker container with the specified configuration exists
func (c *Container) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	container, err := c.client.FindContainer(c.Name)
	if err != nil {
		status.Level = resource.StatusFatal
		return status, err
	}

	if container != nil {
		status.AddDifference("name", strings.TrimPrefix(container.Name, "/"), c.Name, "")
		if c.Force {
			if diffErr := c.diffContainer(container, status); diffErr != nil {
				return nil, diffErr
			}
		}
	} else {
		status.AddDifference("name", "", c.Name, "<container-missing>")
	}

	status.RaiseLevelForDiffs()

	return status, nil
}

// Apply starts a docker container with the specified configuration
func (c *Container) Apply(context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	volumes, binds := volumeConfigs(c.Volumes)
	config := &dc.Config{
		Image:        c.Image,
		WorkingDir:   c.WorkingDir,
		Env:          c.Env,
		ExposedPorts: toPortMap(c.Expose),
		Volumes:      volumes,
		Cmd:          c.Command,
		Entrypoint:   c.Entrypoint,
	}

	hostConfig := &dc.HostConfig{
		PublishAllPorts: c.PublishAllPorts,
		Links:           c.Links,
		DNS:             c.DNS,
		PortBindings:    toPortBindingMap(c.PortBindings),
		Binds:           binds,
		VolumesFrom:     c.VolumesFrom,
		NetworkMode:     c.NetworkMode,
	}

	opts := dc.CreateContainerOptions{
		Name:       c.Name,
		Config:     config,
		HostConfig: hostConfig,
	}

	container, err := c.client.CreateContainer(opts)
	if err != nil {
		return status, err
	}

	for _, name := range c.Networks {
		err = c.client.ConnectNetwork(name, container)
		if err != nil {
			return status, err
		}
	}

	if c.CStatus == "" || c.CStatus == containerStatusRunning {
		err = c.client.StartContainer(c.Name, container.ID)
		if err != nil {
			return status, err
		}
	}

	return status, nil
}

// SetClient injects a docker api client
func (c *Container) SetClient(client docker.APIClient) {
	c.client = client
}

func (c *Container) diffContainer(container *dc.Container, status *resource.Status) error {
	expectedStatus := strings.ToLower(c.CStatus)
	if expectedStatus == "" {
		expectedStatus = containerStatusRunning
	}
	status.AddDifference("status", strings.ToLower(container.State.Status), expectedStatus, "")

	if container.HostConfig != nil {
		status.AddDifference(
			"publish_all_ports",
			strconv.FormatBool(container.HostConfig.PublishAllPorts),
			strconv.FormatBool(c.PublishAllPorts),
			"false")
		status.AddDifference(
			"dns",
			strings.Join(container.HostConfig.DNS, ", "),
			strings.Join(c.DNS, ", "),
			"")
		status.AddDifference(
			"volumes_from",
			strings.Join(container.HostConfig.VolumesFrom, ", "),
			strings.Join(c.VolumesFrom, ", "),
			"")
		status.AddDifference(
			"network_mode",
			container.HostConfig.NetworkMode,
			c.NetworkMode,
			DefaultNetworkMode,
		)
	}

	image, err := c.client.FindImage(container.Image)
	if err != nil {
		status.Level = resource.StatusFatal
		return errors.Wrapf(err, "failed to find image %s for container %s", container.Image, container.Name)
	}
	if image == nil {
		return errors.New("backing image is unavailable")
	}

	var actual, expected string
	// if Cmd is empty, compare using the default from the container Image
	actual = strings.Join(container.Config.Cmd, " ")
	if len(c.Command) == 0 {
		expected = strings.Join(image.Config.Cmd, " ")
	} else {
		expected = strings.Join(c.Command, " ")
	}
	status.AddDifference("command", actual, expected, "")

	// if Entrypoint is empty, compare using the default from the container Image
	actual = strings.Join(container.Config.Entrypoint, " ")
	if len(c.Entrypoint) == 0 {
		expected = strings.Join(image.Config.Entrypoint, " ")
	} else {
		expected = strings.Join(c.Entrypoint, " ")
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

	// Env
	actual, expected = c.compareEnv(container, image)
	status.AddDifference("env", actual, expected, "")

	// Ports
	actual, expected = c.comparePortMappings(container)
	status.AddDifference("ports", actual, expected, "")

	// Expose
	actual, expected = c.compareExposedPorts(container, image)
	status.AddDifference("expose", actual, expected, "")

	// Links
	actual, expected = c.compareLinks(container)
	status.AddDifference("links", actual, expected, "")

	// Volumes
	actual, expected = c.compareVolumes(container, image)
	status.AddDifference("volumes", actual, expected, "")

	// Binds
	actual, expected = c.compareBinds(container)
	status.AddDifference("binds", actual, expected, "")

	// Networks
	actual, expected = c.compareNetworks(container)
	status.AddDifference("networks", actual, expected, "")

	// Image
	existingRepoTag := preferredRepoTag(c.Image, image)
	status.AddDifference("image", existingRepoTag, c.Image, "")

	return nil
}

func (c *Container) compareNetworks(container *dc.Container) (actual, expected string) {
	if container.NetworkSettings == nil {
		return "", ""
	}

	var containerNetworks []string
	for name := range container.NetworkSettings.Networks {
		var isBuiltin bool
		for _, builtin := range builtinNetworks {
			if strings.EqualFold(name, builtin) {
				isBuiltin = true
				break
			}
		}
		if !isBuiltin {
			containerNetworks = append(containerNetworks, name)
		}
	}
	sort.Strings(containerNetworks)

	expectedNetworks := make([]string, len(c.Networks))
	copy(expectedNetworks, c.Networks)
	sort.Strings(expectedNetworks)

	return strings.Join(containerNetworks, ", "), strings.Join(expectedNetworks, ", ")
}

func (c *Container) compareEnv(container *dc.Container, image *dc.Image) (actual, expected string) {
	varName := func(envvar string) string {
		return strings.ToLower(strings.Split(envvar, "=")[0])
	}

	varNameSet := func(env []string) mapset.Set {
		varnames := make([]string, len(env))
		for i, envvar := range env {
			varnames[i] = varName(envvar)
		}
		return toStringSet(varnames)
	}

	varSet := func(env []string, varNames mapset.Set) mapset.Set {
		set := mapset.NewSet()
		for _, envvar := range env {
			if varNames.Contains(varName(envvar)) {
				set.Add(envvar)
			}
		}
		return set
	}

	wantedVarNames := varNameSet(c.Env)                   // desired var names
	imageVarNames := varNameSet(image.Config.Env)         // var names defined in the image config
	engineVarNames := varNameSet(engineEnvVars)           // var names defined by the engine
	containerVarNames := varNameSet(container.Config.Env) // all var names defined in the running container

	// we want to ignore vars that are injected by the image config or the engine
	// unless they are explicitly set/overridden in the desired state
	compareVarNames := wantedVarNames.Union(containerVarNames.Difference(imageVarNames.Union(engineVarNames)))

	compareSet := varSet(container.Config.Env, compareVarNames)
	expectedSet := varSet(c.Env, compareVarNames)

	actual = joinStringSet(compareSet, " ")
	expected = joinStringSet(expectedSet, " ")

	return actual, expected
}

func (c *Container) compareExposedPorts(container *dc.Container, image *dc.Image) (actual, expected string) {
	toSet := func(portMap map[dc.Port]struct{}) mapset.Set {
		portSlice := make([]string, len(portMap))
		idx := 0
		for port := range portMap {
			portSlice[idx] = docker.NewPort(string(port)).String()
			idx++
		}
		return toStringSet(portSlice)
	}

	containerSet := toSet(container.Config.ExposedPorts)
	imageSet := toSet(image.Config.ExposedPorts)
	expectedSet := toSet(toPortMap(c.Expose)).Union(imageSet)

	actual = joinStringSet(containerSet, ", ")
	expected = joinStringSet(expectedSet, ", ")

	return actual, expected
}

func (c *Container) comparePortMappings(container *dc.Container) (actual, expected string) {
	toBindingsList := func(bindings map[dc.Port][]dc.PortBinding) []string {
		var containerBindings []string
		for port, pbindings := range bindings {
			cport := docker.NewPort(string(port)).String()
			for _, pbinding := range pbindings {
				ip := pbinding.HostIP
				hport := pbinding.HostPort
				if hport != "" {
					hport = strings.Split(hport, "/")[0]
				}
				containerBindings = append(containerBindings, fmt.Sprintf("%s:%s:%s", ip, hport, cport))
			}
		}

		sort.Strings(containerBindings)
		return containerBindings
	}

	if container.HostConfig == nil {
		actual = ""
	} else {
		actual = strings.Join(toBindingsList(container.HostConfig.PortBindings), ", ")
	}

	expected = strings.Join(toBindingsList(toPortBindingMap(c.PortBindings)), ", ")

	return actual, expected
}

func (c *Container) compareLinks(container *dc.Container) (actual, expected string) {
	normalizeLink := func(link string) string {
		// internally links are stored as "/linkedcontainername:/containername/alias"
		parts := strings.Split(link, ":")
		if len(parts) == 1 {
			return strings.TrimPrefix(link, "/")
		}

		var name, alias string
		if strings.HasPrefix(parts[0], "/") {
			_, alias = path.Split(parts[1])
			name = parts[0][1:]
		} else {
			name = parts[0]
			alias = parts[1]
		}

		if strings.EqualFold(name, alias) {
			return name
		}

		return fmt.Sprintf("%s:%s", name, alias)
	}

	normalizedLinks := func(rawlinks []string) []string {
		links := make([]string, len(rawlinks))
		for i, link := range rawlinks {
			links[i] = normalizeLink(link)
		}
		sort.Strings(links)
		return links
	}

	if container.HostConfig == nil {
		actual = ""
	} else {
		actual = strings.Join(normalizedLinks(container.HostConfig.Links), ", ")
	}

	expected = strings.Join(normalizedLinks(c.Links), ", ")

	return actual, expected
}

func (c *Container) compareVolumes(container *dc.Container, image *dc.Image) (actual, expected string) {
	toSet := func(vols map[string]struct{}) mapset.Set {
		vollist := make([]string, len(vols))
		idx := 0
		for vol := range vols {
			vollist[idx] = vol
			idx++
		}
		return toStringSet(vollist)
	}

	containerSet := toSet(container.Config.Volumes)
	imageSet := toSet(image.Config.Volumes)

	volumes, _ := volumeConfigs(c.Volumes)
	expectedSet := toSet(volumes).Union(imageSet)

	actual = joinStringSet(containerSet, ", ")
	expected = joinStringSet(expectedSet, ", ")

	return actual, expected
}

func (c *Container) compareBinds(container *dc.Container) (actual, expected string) {
	toList := func(binds []string) []string {
		list := make([]string, len(binds))
		copy(list, binds)
		sort.Strings(list)
		return list
	}

	if container.HostConfig == nil {
		actual = ""
	} else {
		actual = strings.Join(toList(container.HostConfig.Binds), ", ")
	}

	_, binds := volumeConfigs(c.Volumes)
	expected = strings.Join(toList(binds), ", ")

	return actual, expected
}

func toPortBindingMap(portBindings []string) map[dc.Port][]dc.PortBinding {
	bindings := make(map[dc.Port][]dc.PortBinding)
	for _, mapping := range portBindings {
		parts := strings.Split(mapping, ":")
		partslen := len(parts)
		switch {
		case partslen == 1:
			cport := docker.NewPort(parts[0]).ToDockerClientPort()
			bindings[cport] = append(bindings[cport], dc.PortBinding{})
		case partslen == 2:
			hport := parts[0]
			cport := docker.NewPort(parts[1]).ToDockerClientPort()
			bindings[cport] = append(bindings[cport], dc.PortBinding{HostPort: hport})
		case partslen > 2:
			ip := parts[0]
			hport := parts[1]
			cport := docker.NewPort(parts[2]).ToDockerClientPort()
			bindings[cport] = append(bindings[cport], dc.PortBinding{HostPort: hport, HostIP: ip})
		}
	}
	return bindings
}

func toStringSet(strings []string) mapset.Set {
	set := mapset.NewSet()
	for _, val := range strings {
		set.Add(val)
	}
	return set
}

func joinStringSet(set mapset.Set, sep string) string {
	list := set.ToSlice()
	strlist := make([]string, len(list))
	for i, val := range list {
		strlist[i] = val.(string)
	}
	sort.Strings(strlist)
	return strings.Join(strlist, sep)
}

func toPortMap(portList []string) map[dc.Port]struct{} {
	portMap := make(map[dc.Port]struct{}, len(portList))
	for _, e := range portList {
		port := docker.NewPort(e).ToDockerClientPort()
		portMap[port] = struct{}{}
	}
	return portMap
}

func volumeConfigs(vols []string) (map[string]struct{}, []string) {
	volumes := make(map[string]struct{})
	binds := []string{}
	for _, vol := range vols {
		parts := strings.Split(vol, ":")
		partslen := len(parts)

		switch {
		case partslen == 1:
			volumes[parts[0]] = struct{}{}
		case partslen > 1:
			volume := parts[1]
			volumes[volume] = struct{}{}
			binds = append(binds, vol)
		}
	}

	return volumes, binds
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
