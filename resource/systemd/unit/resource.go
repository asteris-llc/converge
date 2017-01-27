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

package unit

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

type Resource struct {
	Name         string `export:"unit"`
	State        string `export:"state"`
	Reload       bool   `export:"reload"`
	SignalName   string `export:"signal_name"`
	SignalNumber uint   `export:"signal_number"`

	// These values are set automatically at check time and contain information
	// about the current systemd process.  They are used for generating messages
	// and to provide rich exported information about systemd processes.

	// The full path to the unit file on disk
	Path string `export:"path"`

	// Description of the services
	Description string `export:"description"`

	// The active state of the unit
	ActiveState string `export:"activestate"`

	// The load state of the unit
	LoadState string `export:"loadstate"`

	// The type of the unit
	Type UnitType `export:"type"`

	// The status represents the current status of the process.  It will be
	// initialized during planning and updated after apply to reflect any changes.

	Status string `export:"status"`

	Properties          *Properties              `export:"global_properties"`
	ServiceProperties   *ServiceTypeProperties   `export:"service_properties"`
	SocketProperties    *SocketTypeProperties    `export:"SocketProperties"`
	DeviceProperties    *DeviceTypeProperties    `export:"DeviceProperties"`
	MountProperties     *MountTypeProperties     `export:"MountProperties"`
	AutomountProperties *AutomountTypeProperties `export:"AutomountProperties"`
	SwapProperties      *SwapTypeProperties      `export:"SwapProperties"`
	PathProperties      *PathTypeProperties      `export:"PathProperties"`
	TimerProperties     *TimerTypeProperties     `export:"TimerProperties"`
	SliceProperties     *SliceTypeProperties     `export:"SliceProperties"`
	ScopeProperties     *ScopeTypeProperties     `export:"ScopeProperties"`

	sendSignal      bool
	systemdExecutor SystemdExecutor
}

func (r *Resource) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	u, err := r.systemdExecutor.QueryUnit(r.Name, true)

	if err != nil {
		return nil, fmt.Errorf("No unit named '%s'. Is it loaded?", r.Name)
	}

	r.populateFromUnit(u)

	if r.sendSignal {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage(fmt.Sprintf("Sending signal `%s` to process", r.SignalName))
	}

	if r.Reload {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage("Reloading unit configuration")
		status.AddDifference("state", u.ActiveState, "reloaded", "")
	}

	if r.State == "restared" {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddMessage("Restarting unit")
		status.AddDifference("state", u.ActiveState, "restarted", "")
	}

	return status, nil
}

func (r *Resource) Apply(context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	return status, nil
}

// We copy data from the unit into the resource to make the UX nicer for users
// who want to access systemd information.
func (r *Resource) populateFromUnit(u *Unit) {
	r.Description = u.Description
	r.Path = u.Path
	r.Type = u.Type
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
		reason := getFailedReason(u)
		st.AddMessage("unit has failed; will attempt to restart")
		st.AddMessage(fmt.Sprintf("unit previously failed, the message was: %s", reason))
		st.RaiseLevel(resource.StatusWillChange)
		st.AddDifference("state", "failed", "active", "")
		return true
	case "activating":
		st.AddMessage("unit is alread activating, will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		return false
	case "deactivating":
		st.AddMessage("unit is currently deactivating, will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		st.AddDifference("state", "deactivating", "active", "")
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
		st.AddMessage("unit failed due to: " + getFailedReason(u))
		return false
	case "activating":
		st.AddDifference("state", "active", "inactive", "")
		st.RaiseLevel(resource.StatusMayChange)
		return true
	case "deactivating":
		st.AddMessage("unit is deactivating.  Will re-check status during apply")
		st.RaiseLevel(resource.StatusMayChange)
		return false
	}
	return true
}

func getFailedReason(u *Unit) string {
	var reason string
	switch u.Type {
	case UnitTypeService:
		reason = u.ServiceProperties.Result
	case UnitTypeSocket:
		reason = u.SocketProperties.Result
	case UnitTypeMount:
		reason = u.MountProperties.Result
	case UnitTypeAutoMount:
		reason = u.MountProperties.Result
	case UnitTypeSwap:
		reason = u.SwapProperties.Result
	case UnitTypeTimer:
		reason = u.TimerProperties.Result
	}
	switch reason {
	case "success":
		return "the unit was activated successfully"
	case "resources":
		return "not enough resources available to create the process"
	case "timeout":
		return "a timeout occurred while starting the unit"
	case "exit-code":
		return "unit exited with a non-zero exit code"
	case "signal":
		return "unit exited due to an unhandled signal"
	case "core-dump":
		return "unit exited and dumped core"
	case "watchdog":
		return "watchdog terminated the service due to slow or missing responses"
	case "start-limit":
		return "process has been restarted too many times"
	case "service-failed-permanent":
		return "continual failure of this socket"
	}
	return "unkown reason"
}
