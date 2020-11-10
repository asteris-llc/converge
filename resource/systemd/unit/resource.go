// Copyright © 2017 Asteris, LLC
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

package unit

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Resource is the resource struct for systemd.unit.
//
// Many of the exported fields are derived from the systemd dbus api; see
// https://www.freedesktop.org/wiki/Software/systemd/dbus/ for a full
// description of the potential values for this fields and their meanings.
type Resource struct {
	// The name of the unit, including the unit type.
	Name string `export:"unit"`

	// The desired state of the unit as configured by the user. It will be one of
	// `running`, `stopped`, or `restarted` if it was configured by the user, and
	// an empty string otherwise.
	State string `export:"state"`

	// This field is set to true if the reload flag was configured by the user.
	Reload bool `export:"reload"`

	// The human-readable name of a unix signal that will be sent to the process.
	// If this is set the name will match the field set in SignalNumber.  See the
	// man pages for `signal(3)` on BSD/Darwin or `signal(7)` on GNU Linux for a
	// full explanation of these signals.
	SignalName string `export:"signal_name"`

	// The numeric identifier of a unix signal that will be sent to the process.
	// If this is set it will match the field set in SignalName.  See the man
	// pages for `signal(3)` on BSD/Darwin or `signal(7)` on GNU Linux for a full
	// explanation of these signals.
	SignalNumber uint `export:"signal_number"`

	// Set to true if the unit is enabled (symlinks exist in `/etc`)
	Enabled bool `export:"enabled"`

	// Set to tru if the unit is enabled for runtime (symlinks exist in `/run`)
	EnabledRuntime bool `export:"enabled_runtime"`

	// The full path to the unit file on disk. This field will be empty if the
	// unit was not started from a systemd unit file on disk.
	Path string `export:"path"`

	// Description of the services. This field will be empty unless a description
	// has been provided in the systemd unit file.
	Description string `export:"description"`

	// The active state of the unit. It will always be one of: `active`,
	// `reloading`, `inactive`, `failed`, `activating`, `deactivating`.
	ActiveState string `export:"activestate"`

	// The load state of the unit.
	LoadState string `export:"loadstate"`

	// The type of the unit as an enumerated value.  See TypeStr for a human
	// readable type.
	Type UnitType `export:"type"`

	// The type of the unit as a human readable string.  See the man page for
	// `systemd(1)` for a full description of the types and their meaning.
	TypeStr string `export:"typestr"`

	// The status represents the current status of the process.  It will be
	// initialized during planning and updated after apply to reflect any changes.
	Status string `export:"status"`

	// Properties are the global systemd unit properties and will be set for all
	// unit types. See the [systemd_Properties]({{< ref "properties.md" >}}) docs
	// for more information.
	Properties Properties `re-export-as:"global_properties"`

	// ServiceProperties contain properties specific to Service unit types. This
	// field is only exported when the unit type is `service`. See the
	// [systemd_ServiceTypeProperties]({{< ref "service_properties.md" >}}) docs
	// for more information.
	ServiceProperties *ServiceTypeProperties `re-export-as:"service_properties"`

	// SocketProperties contain properties specific to Socket unit types. This
	// field is only exported when the unit type is `socket`. See the
	// [systemd_SocketTypeProperties]({{< ref "socket_properties.md" >}}) docs for
	// more information.
	SocketProperties *SocketTypeProperties `re-export-as:"socket_properties"`

	// DeviceProperties contain properties specific to Device unit types. This
	// field is only exported when the unit type is `device`. See the
	// [systemd_DeviceTypeProperties]({{< ref "device_properties.md" >}}) docs for
	// more information.
	DeviceProperties *DeviceTypeProperties `re-export-as:"device_properties"`

	// MountProperties contain properties specific to Mount unit types. This field
	// is only exported when the unit type is `mount`. See the
	// [systemd_MountTypeProperties]({{< ref "mount_properties.md" >}}) docs for
	// more information.
	MountProperties *MountTypeProperties `re-export-as:"mount_properties"`

	// AutomountProperties contain properties specific to Autoumount unit types.
	// This field is only exported when the unit type is`automount`. See the
	// [systemd_AutomountTypeProperties]({{< ref "automount_properties.md" >}})
	// docs for more information.
	AutomountProperties *AutomountTypeProperties `re-export-as:"automount_properties"`

	// SwapProperties contain properties specific to Swap unit types. This field
	// is only exported when the unit type is `swap`. See the
	// [systemd_SwapTypeProperties]({{< ref "swap_properties.md" >}}) docs for
	// more information.
	SwapProperties *SwapTypeProperties `re-export-as:"swap_properties"`

	// PathProperties contain properties specific to Path unit types. This field
	// is only exported when the unit type is `path`. See the
	// [systemd_PathTypeProperties]({{< ref "path_properties.md" >}}) docs for
	// more information.
	PathProperties *PathTypeProperties `re-export-as:"path_properties"`

	// TimerProperties contain properties specific to Timer unit types. This field
	// is only exported when the unit type is `timer`. See the
	// [systemd_TimerTypeProperties]({{< ref "timer_properties.md" >}}) docs for
	// more information.
	TimerProperties *TimerTypeProperties `re-export-as:"timer_properties"`

	// SliceProperties contain properties specific to Slice unit types. This field
	// is only exported when the unit type is `slice`. See the
	// [systemd_SliceTypeProperties]({{< ref "slice_properties.md" >}}) docs for
	// more information.
	SliceProperties *SliceTypeProperties `re-export-as:"slice_properties"`

	// ScopeProperties contain properties specific to Scope unit types. This field
	// is only exported when the unit type is `scope`. See the
	// [systemd_ScopeTypeProperties]({{< ref "scope_properties.md" >}}) docs for
	// more information.
	ScopeProperties *ScopeTypeProperties `re-export-as:"scope_properties"`

	enableChange        *bool // enabled if true, disabled if false, unmodified if nil
	enableRuntimeChange *bool // enabled if true, disabled if false, unmodified if nil
	sendSignal          bool
	systemdExecutor     SystemdExecutor
	hasRun              bool
	fs                  fsexecutor
}

type response struct {
	status resource.TaskStatus
	err    error
}

func wrapCall(f func() (resource.TaskStatus, error)) <-chan response {
	resp := make(chan response)
	go func() {
		st, err := f()
		resp <- response{st, err}
	}()
	return resp
}

// Check implements resource.Task
func (r *Resource) Check(ctx context.Context, _ resource.Renderer) (resource.TaskStatus, error) {
	ch := wrapCall(r.runCheck)
	select {
	case <-ctx.Done():
		return nil, errors.New("context was cancelled")
	case results := <-ch:
		return results.status, results.err
	}
}

// Apply implemnts resource.Task
func (r *Resource) Apply(ctx context.Context) (resource.TaskStatus, error) {
	ch := wrapCall(r.runApply)
	select {
	case <-ctx.Done():
		return nil, errors.New("context was cancelled")
	case results := <-ch:
		return results.status, results.err
	}
}

func (r *Resource) runCheck() (resource.TaskStatus, error) {
	status := resource.NewStatus()
	u, err := r.systemdExecutor.QueryUnit(r.Name, false)
	if err != nil {
		status.AddMessage("query unit returned an error: " + err.Error())
		return nil, err
	}
	r.populateFromUnit(u)
	if r.sendSignal && !r.hasRun {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage(fmt.Sprintf("Sending signal `%s` to unit", r.SignalName))
	}
	if r.Reload && !r.hasRun {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage("Reloading unit configuration")
		status.AddDifference("state", u.ActiveState, "reloaded", "")
	}
	switch r.State {
	case "restarted":
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage("Restarting unit")
		status.AddDifference("state", u.ActiveState, "restarted", "")
	case "running":
		r.shouldStart(u, status)
	case "stopped":
		r.shouldStop(u, status)
	}
	enabledRuntime, enabledPersistent, err := r.isEnabled(u)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine if the unit is enabled")
	}
	r.Enabled = enabledPersistent
	r.EnabledRuntime = enabledRuntime

	status = r.updateEnableStatus("persistent",
		status,
		r.Enabled,
		r.enableChange)
	status = r.updateEnableStatus("runtime",
		status,
		r.EnabledRuntime,
		r.enableRuntimeChange)
	r.hasRun = true
	return status, nil
}

func (r *Resource) updateEnableStatus(
	msg string,
	status *resource.Status,
	current bool,
	want *bool) *resource.Status {
	var tgt bool
	if want == nil {
		return status
	}
	tgt = *want
	if tgt == current {
		status.AddMessage(fmt.Sprintf("%s unit is already %s", msg, showEnabled(tgt)))
		return status
	}
	status.RaiseLevel(resource.StatusWillChange)
	status.AddDifference(msg, showEnabled(current), showEnabled(tgt), "")
	return status
}

func showEnabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func (r *Resource) runApply() (resource.TaskStatus, error) {
	status := resource.NewStatus()
	tempStatus := resource.NewStatus()
	u, err := r.systemdExecutor.QueryUnit(r.Name, false)
	if err != nil {
		return nil, err
	}

	if u.ActiveState == "unknown" {
		log.Infof("unable to query information about the unit. " +
			"Making a best guess based on configured data")
		u.Name = r.Name
	}

	if r.sendSignal {
		status.AddMessage(fmt.Sprintf("Sending signal `%s` to unit", r.SignalName))
		r.systemdExecutor.SendSignal(u, Signal(r.SignalNumber))
	}
	if r.Reload {
		status.AddMessage("Reloading unit configuration")
		status.AddDifference("state", u.ActiveState, "reloaded", "")
		if err := r.systemdExecutor.ReloadUnit(u); err != nil {
			return nil, err
		}
	}

	var runstateErr error
	switch r.State {
	case "running":
		if r.shouldStart(u, tempStatus) {
			runstateErr = r.systemdExecutor.StartUnit(u)
		}
	case "stopped":
		if r.shouldStop(u, tempStatus) {
			runstateErr = r.systemdExecutor.StopUnit(u)
		}
	case "restarted":
		runstateErr = r.systemdExecutor.RestartUnit(u)
	}

	enabledRuntime, enabledPersistent, err := r.isEnabled(u)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine if the unit is enabled")
	}
	r.Enabled = enabledPersistent
	r.EnabledRuntime = enabledRuntime

	var symlinkChanges []*unitFileChange

	if r.enableChange != nil {
		changes, err := r.toggleUnitEnabled(u, status, false, *r.enableChange, enabledPersistent)
		if err != nil {
			return status, err
		}
		symlinkChanges = append(symlinkChanges, changes...)
	}

	if r.enableRuntimeChange != nil {
		changes, err := r.toggleUnitEnabled(u, status, true, *r.enableRuntimeChange, enabledRuntime)
		if err != nil {
			return status, err
		}
		symlinkChanges = append(symlinkChanges, changes...)
	}

	for _, ch := range symlinkChanges {
		status.AddDifference(ch.Type, "", fmt.Sprintf("%s -> %s", ch.Filename, ch.Destination), "")
	}

	return status, runstateErr
}

func (r *Resource) toggleUnitEnabled(u *Unit, status *resource.Status, runtime, shouldBeEnabled, isEnabled bool) ([]*unitFileChange, error) {
	if shouldBeEnabled == isEnabled {
		if isEnabled {
			status.AddMessage("unit is already enabled")
		} else {
			status.AddMessage("unit is already disabled")
		}
		return []*unitFileChange{}, nil
	}
	if shouldBeEnabled {
		_, c, e := r.systemdExecutor.EnableUnit(u, runtime, true)
		return c, e
	}
	return r.systemdExecutor.DisableUnit(u, runtime)
}

// We copy data from the unit into the resource to make the UX nicer for users
// who want to access systemd information.
func (r *Resource) populateFromUnit(u *Unit) {
	r.Description = u.Description
	r.Path = u.Path
	r.Type = u.Type
	r.TypeStr = u.Type.String()
	r.Status = u.ActiveState
	r.Properties = u.Properties
	r.ServiceProperties = u.ServiceProperties
	r.SocketProperties = u.SocketProperties
	r.DeviceProperties = u.DeviceProperties
	r.MountProperties = u.MountProperties
	r.AutomountProperties = u.AutomountProperties
	r.SwapProperties = u.SwapProperties
	r.PathProperties = u.PathProperties
	r.TimerProperties = u.TimerProperties
	r.SliceProperties = u.SliceProperties
	r.ScopeProperties = u.ScopeProperties
}

func (r *Resource) shouldStart(u *Unit, st *resource.Status) bool {
	switch u.ActiveState {
	case "active":
		st.AddMessage("already running")
		return false
	case "reloading":
		st.AddMessage("unit is reloading, will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "inactive":
		st.RaiseLevel(resource.StatusWillChange)
		st.AddDifference("state", "inactive", "active", "")
		return true
	case "failed":
		st.AddMessage("unit has failed; will attempt to restart")
		if reason, err := getFailedReason(u); err != nil {
			st.AddMessage(fmt.Sprintf("cannot determine root cause of failure: %v", err))
		} else {
			st.AddMessage(fmt.Sprintf("unit previously failed, the message was: %s", reason))
		}
		st.RaiseLevel(resource.StatusWillChange)
		st.AddDifference("state", "failed", "active", "")
		return true
	case "activating":
		st.AddMessage("unit is alread activating, will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "deactivating":
		st.AddMessage("unit is currently deactivating, will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		st.AddDifference("state", "deactivating", "active", "")
		return true
	case "unknown":
		st.AddMessage("unit was in an unknown state, will attempt to start")
		st.RaiseLevel(resource.StatusMayChange)
		st.AddDifference("state", "unknown", "active", "")
		return true
	}
	return true
}

func (r *Resource) shouldStop(u *Unit, st *resource.Status) bool {
	switch u.ActiveState {
	case "active":
		st.AddDifference("state", "active", "inactive", "")
		st.RaiseLevel(resource.StatusWillChange)
		return true
	case "reloading":
		st.AddMessage("unit is reloading; will re-check status during apply")
		st.AddDifference("state", "reloading", "inactive", "")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "inactive":
		st.AddMessage("unit is already inactive")
		return false
	case "failed":
		st.AddMessage("unit is not running because it has failed.  Will not restart")
		if reason, err := getFailedReason(u); err != nil {
			st.AddMessage(fmt.Sprintf("cannot determine root cause of failure: %v", err))
		} else {
			st.AddMessage(fmt.Sprintf("unit previously failed, the message was: %s", reason))
		}
		return false
	case "activating":
		st.AddDifference("state", "active", "inactive", "")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "deactivating":
		st.AddMessage("unit is deactivating.  Will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "unknown":
		st.AddMessage("unit was in an unknown state, will attempt to stop")
		st.RaiseLevel(resource.StatusMayChange)
		st.AddDifference("state", "unknown", "inactive", "")
		return true
	}
	return true
}

// isEnabled checks a unit file to see if it's enabled at runtime and/or
// persistently. It returns a thruple of the runtime enablement, system
// enablement, and an error
func (r *Resource) isEnabled(unit *Unit) (runtime bool, persistent bool, err error) {
	runtime, err = r.existsInTree("/run/systemd", unit)
	if err != nil {
		return false, false, err
	}
	persistent, err = r.existsInTree("/etc/systemd", unit)
	if err != nil {
		return false, false, err
	}
	return runtime, persistent, nil
}

func (r *Resource) existsInTree(root string, unit *Unit) (bool, error) {
	var found bool
	toFind := unit.Name
	var checkSymlink bool

	fmt.Fprintf(os.Stderr, "existsInTree (fprintf to stderr)\n")

	if unit == nil {
		fmt.Fprintf(os.Stderr, "unit is nil in existsInTree!")
		return false, errors.New("unit is nil")
	}

	if unit.Path != "" {
		toFind = unit.Path
		checkSymlink = true
	}

	err := r.fs.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if checkSymlink {
			matches, matchErr := r.symlinkTargetMatches(path, toFind)
			found = found || matches
			err = matchErr
		} else if info.Name() == toFind {
			found = true
		}
		return nil
	})
	return found, err
}

func (r *Resource) symlinkTargetMatches(symlinkPath, expectedPath string) (bool, error) {
	canonical, err := r.fs.EvalSymlinks(symlinkPath)
	if err != nil {
		return false, err
	}
	return (canonical == expectedPath), nil
}

func getFailedReason(u *Unit) (string, error) {
	err := errors.New("unable to determine cause of failure: no properties available")
	var reason string
	switch u.Type {
	case UnitTypeService:
		if u.ServiceProperties == nil {
			return "", err
		}
		reason = u.ServiceProperties.Result
	case UnitTypeSocket:
		if u.SocketProperties == nil {
			return "", err
		}
		reason = u.SocketProperties.Result
	case UnitTypeMount:
		if u.MountProperties == nil {
			return "", err
		}
		reason = u.MountProperties.Result
	case UnitTypeAutoMount:
		if u.AutomountProperties == nil {
			return "", err
		}
		reason = u.AutomountProperties.Result
	case UnitTypeSwap:
		if u.SwapProperties == nil {
			return "", err
		}
		reason = u.SwapProperties.Result
	case UnitTypeTimer:
		if u.TimerProperties == nil {
			return "", err
		}
		reason = u.TimerProperties.Result
	}
	switch reason {
	case "success":
		return "the unit was activated successfully", nil
	case "resources":
		return "not enough resources available to create the process", nil
	case "timeout":
		return "a timeout occurred while starting the unit", nil
	case "exit-code":
		return "unit exited with a non-zero exit code", nil
	case "signal":
		return "unit exited due to an unhandled signal", nil
	case "core-dump":
		return "unit exited and dumped core", nil
	case "watchdog":
		return "watchdog terminated the service due to slow or missing responses", nil
	case "start-limit":
		return "process has been restarted too many times", nil
	case "service-failed-permanent":
		return "continual failure of this socket", nil
	}
	return "unknown reason", nil
}
