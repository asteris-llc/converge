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

package unit_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/asteris-llc/converge/resource/systemd/unit"
	"github.com/stretchr/testify/assert"
)

func TestTaskInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(unit.Unit))
}

func TestInactiveToActiveUnit(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	foo := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	status, err := foo.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled\", expected one of [\"enabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestDisabledtoEnabledUnit(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply(&fr)
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	status, err := enabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled\", expected one of [\"enabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestDisabledtoEnabledRuntimeUnit(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply(&fr)
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled-runtime", StartMode: "replace"}
	status, err := enabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled-runtime\", expected one of [\"enabled-runtime, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestEnabledToDisabledUnit(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	_, err = enabled.Apply(&fr)
	disabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "disabled", StartMode: "replace"}
	status, err := disabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"disabled\", expected one of [\"disabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestInactiveToActiveUnitEtc(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/etc", false)
	assert.NoError(t, err)
	defer svc.Remove()

	foo := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	status, err := foo.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled\", expected one of [\"enabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestDisabledtoEnabledUnitEtc(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/etc", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply(&fr)
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	status, err := enabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled\", expected one of [\"enabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestDisabledtoEnabledRuntimeUnitEtc(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/etc", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply(&fr)
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled-runtime", StartMode: "replace"}
	status, err := enabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled-runtime\", expected one of [\"enabled-runtime, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

func TestEnabledToDisabledUnitEtc(t *testing.T) {
	// t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/etc", false)
	assert.NoError(t, err)
	defer svc.Remove()

	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	_, err = enabled.Apply(&fr)
	disabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "disabled", StartMode: "replace"}
	status, err := disabled.Apply(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("property \"UnitFileState\" of unit %q is \"disabled\", expected one of [\"disabled, static\"]", svc.Name), status.Value())
	assert.False(t, status.HasChanges())
}

const HelloUnit = `
[Unit]
Description=Foo hello world
[Service]
ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"

[Install]
WantedBy=multi-user.target
`

const HelloUnitStatic = `
[Unit]
Description=Foo hello world
[Service]
ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"
`

type TmpService struct {
	Path string
	Name string
}

func (t *TmpService) Remove() {
	base := filepath.Base(t.Path)

	disable := "systemctl disable " + base
	daemonReload := "systemctl daemon-reload"
	resetFailed := "systemctl reset-failed"

	generator := &shell.CommandGenerator{Interpreter: "/bin/bash"}
	generator.Run(disable)

	os.Remove(t.Path)
	locations := []string{"/run/systemd/system/", "/run/systemd/system/", "/usr/lib/systemd/system/"}
	for _, l := range locations {
		os.Remove(filepath.Join(l, base))
	}

	generator.Run(daemonReload)
	generator.Run(resetFailed)
}

var count uint32

func NewTmpService(prefix string, static bool) (svc *TmpService, err error) {
	atomic.AddUint32(&count, 1)
	name := fmt.Sprintf("foo%d.service", count)
	path := filepath.Join(prefix, "systemd/system", name)
	if static {
		err = ioutil.WriteFile(path, []byte(HelloUnitStatic), 0777)
	} else {
		err = ioutil.WriteFile(path, []byte(HelloUnit), 0777)
	}
	if err != nil {
		return nil, err
	}
	return &TmpService{Path: path, Name: name}, systemd.ApplyDaemonReload()
}

func IsRoot() bool {
	currentUser, _ := user.Current()
	return currentUser.Uid == "0"
}

func HasSystemd() bool {
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return false
	}
	defer conn.Return()
	return true
}
