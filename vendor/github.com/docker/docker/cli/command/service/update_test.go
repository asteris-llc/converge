package service

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	mounttypes "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestUpdateServiceArgs(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("args", "the \"new args\"")

	spec := &swarm.ServiceSpec{}
	cspec := &spec.TaskTemplate.ContainerSpec
	cspec.Args = []string{"old", "args"}

	updateService(flags, spec)
	assert.EqualStringSlice(t, cspec.Args, []string{"the", "new args"})
}

func TestUpdateLabels(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("label-add", "toadd=newlabel")
	flags.Set("label-rm", "toremove")

	labels := map[string]string{
		"toremove": "thelabeltoremove",
		"tokeep":   "value",
	}

	updateLabels(flags, &labels)
	assert.Equal(t, len(labels), 2)
	assert.Equal(t, labels["tokeep"], "value")
	assert.Equal(t, labels["toadd"], "newlabel")
}

func TestUpdateLabelsRemoveALabelThatDoesNotExist(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("label-rm", "dne")

	labels := map[string]string{"foo": "theoldlabel"}
	updateLabels(flags, &labels)
	assert.Equal(t, len(labels), 1)
}

func TestUpdatePlacement(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("constraint-add", "node=toadd")
	flags.Set("constraint-rm", "node!=toremove")

	placement := &swarm.Placement{
		Constraints: []string{"node!=toremove", "container=tokeep"},
	}

	updatePlacement(flags, placement)
	assert.Equal(t, len(placement.Constraints), 2)
	assert.Equal(t, placement.Constraints[0], "container=tokeep")
	assert.Equal(t, placement.Constraints[1], "node=toadd")
}

func TestUpdateEnvironment(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-add", "toadd=newenv")
	flags.Set("env-rm", "toremove")

	envs := []string{"toremove=theenvtoremove", "tokeep=value"}

	updateEnvironment(flags, &envs)
	assert.Equal(t, len(envs), 2)
	// Order has been removed in updateEnvironment (map)
	sort.Strings(envs)
	assert.Equal(t, envs[0], "toadd=newenv")
	assert.Equal(t, envs[1], "tokeep=value")
}

func TestUpdateEnvironmentWithDuplicateValues(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-add", "foo=newenv")
	flags.Set("env-add", "foo=dupe")
	flags.Set("env-rm", "foo")

	envs := []string{"foo=value"}

	updateEnvironment(flags, &envs)
	assert.Equal(t, len(envs), 0)
}

func TestUpdateEnvironmentWithDuplicateKeys(t *testing.T) {
	// Test case for #25404
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-add", "A=b")

	envs := []string{"A=c"}

	updateEnvironment(flags, &envs)
	assert.Equal(t, len(envs), 1)
	assert.Equal(t, envs[0], "A=b")
}

func TestUpdateGroups(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("group-add", "wheel")
	flags.Set("group-add", "docker")
	flags.Set("group-rm", "root")
	flags.Set("group-add", "foo")
	flags.Set("group-rm", "docker")

	groups := []string{"bar", "root"}

	updateGroups(flags, &groups)
	assert.Equal(t, len(groups), 3)
	assert.Equal(t, groups[0], "bar")
	assert.Equal(t, groups[1], "foo")
	assert.Equal(t, groups[2], "wheel")
}

func TestUpdateMounts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("mount-add", "type=volume,target=/toadd")
	flags.Set("mount-rm", "/toremove")

	mounts := []mounttypes.Mount{
		{Target: "/toremove", Type: mounttypes.TypeBind},
		{Target: "/tokeep", Type: mounttypes.TypeBind},
	}

	updateMounts(flags, &mounts)
	assert.Equal(t, len(mounts), 2)
	assert.Equal(t, mounts[0].Target, "/tokeep")
	assert.Equal(t, mounts[1].Target, "/toadd")
}

func TestUpdatePorts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "1000:1000")
	flags.Set("publish-rm", "333/udp")

	portConfigs := []swarm.PortConfig{
		{TargetPort: 333, Protocol: swarm.PortConfigProtocolUDP},
		{TargetPort: 555},
	}

	err := updatePorts(flags, &portConfigs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(portConfigs), 2)
	// Do a sort to have the order (might have changed by map)
	targetPorts := []int{int(portConfigs[0].TargetPort), int(portConfigs[1].TargetPort)}
	sort.Ints(targetPorts)
	assert.Equal(t, targetPorts[0], 555)
	assert.Equal(t, targetPorts[1], 1000)
}

func TestUpdatePortsDuplicateEntries(t *testing.T) {
	// Test case for #25375
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "80:80")

	portConfigs := []swarm.PortConfig{
		{TargetPort: 80, PublishedPort: 80},
	}

	err := updatePorts(flags, &portConfigs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(portConfigs), 1)
	assert.Equal(t, portConfigs[0].TargetPort, uint32(80))
}

func TestUpdatePortsDuplicateKeys(t *testing.T) {
	// Test case for #25375
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "80:20")

	portConfigs := []swarm.PortConfig{
		{TargetPort: 80, PublishedPort: 80},
	}

	err := updatePorts(flags, &portConfigs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(portConfigs), 1)
	assert.Equal(t, portConfigs[0].TargetPort, uint32(20))
}

func TestUpdatePortsConflictingFlags(t *testing.T) {
	// Test case for #25375
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "80:80")
	flags.Set("publish-add", "80:20")

	portConfigs := []swarm.PortConfig{
		{TargetPort: 80, PublishedPort: 80},
	}

	err := updatePorts(flags, &portConfigs)
	assert.Error(t, err, "conflicting port mapping")
}

func TestUpdateHealthcheckTable(t *testing.T) {
	type test struct {
		flags    [][2]string
		initial  *container.HealthConfig
		expected *container.HealthConfig
		err      string
	}
	testCases := []test{
		{
			flags:    [][2]string{{"no-healthcheck", "true"}},
			initial:  &container.HealthConfig{Test: []string{"CMD-SHELL", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Test: []string{"NONE"}},
		},
		{
			flags:    [][2]string{{"health-cmd", "cmd1"}},
			initial:  &container.HealthConfig{Test: []string{"NONE"}},
			expected: &container.HealthConfig{Test: []string{"CMD-SHELL", "cmd1"}},
		},
		{
			flags:    [][2]string{{"health-retries", "10"}},
			initial:  &container.HealthConfig{Test: []string{"NONE"}},
			expected: &container.HealthConfig{Retries: 10},
		},
		{
			flags:    [][2]string{{"health-retries", "10"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
		},
		{
			flags:    [][2]string{{"health-interval", "1m"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Interval: time.Minute},
		},
		{
			flags:    [][2]string{{"health-cmd", ""}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Retries: 10},
		},
		{
			flags:    [][2]string{{"health-retries", "0"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
		},
		{
			flags: [][2]string{{"health-cmd", "cmd1"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
		{
			flags: [][2]string{{"health-interval", "10m"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
		{
			flags: [][2]string{{"health-timeout", "1m"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
	}
	for i, c := range testCases {
		flags := newUpdateCommand(nil).Flags()
		for _, flag := range c.flags {
			flags.Set(flag[0], flag[1])
		}
		cspec := &swarm.ContainerSpec{
			Healthcheck: c.initial,
		}
		err := updateHealthcheck(flags, cspec)
		if c.err != "" {
			assert.Error(t, err, c.err)
		} else {
			assert.NilError(t, err)
			if !reflect.DeepEqual(cspec.Healthcheck, c.expected) {
				t.Errorf("incorrect result for test %d, expected health config:\n\t%#v\ngot:\n\t%#v", i, c.expected, cspec.Healthcheck)
			}
		}
	}
}
