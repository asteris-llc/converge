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

	// The name of the process that is executing
	PSName string `export:"process_name"`

	// The PID of the current job
	Pid string `export:"pid"`

	// Description of the services
	Description string `export:"description"`

	// The status represents the current status of the process.  It will be
	// initialized during planning and updated after apply to reflect any changes.

	Status string `export:"status"`
	Type   string `export:"type"`

	sendSignal bool

	systemdExecutor SystemdExecutor
}

func (r *Resource) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	return status, nil
}

func (r *Resource) Apply(context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	return status, nil
}
