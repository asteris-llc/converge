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

package enable

import (
	"context"
	"fmt"
	"time"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Enable uses dbus to enable a unit
type Enable struct {
	Unit    string //TODO when arrays are implemented, change this to array
	Runtime bool   //Unit enabled for runtime only (true, run), or perstently (false, /etc)
	Force   bool   //whether symlinks pointing to other units shall be replaced if necessary.
	Timeout time.Duration
}

// Check if unit is enabled
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the active state
func (t *Enable) Check() (resource.TaskStatus, error) {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	//Create context
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if t.Timeout == 0 {
		t.Timeout = systemd.DefaultTimeout
	}
	ctx, cancel = context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	// Wait for the connection to have loaded
	err = systemd.WaitToLoad(ctx, dbusConn, t.Unit)
	if err != nil {
		return nil, err
	}
	status, err := systemd.CheckUnitIsActive(dbusConn, t.Unit)
	if err != nil {
		return status, err
	}
	//Check runtime
	if t.Runtime {
		ufsRuntimeStatus, e := systemd.CheckUnitHasEnabledUFSRuntimes(dbusConn, t.Unit)
		ufsRuntimeStatus.Merge(status)
		status = ufsRuntimeStatus
		err = helpers.MultiErrorAppend(e, err)
	} else {
		ufsStatus, e := systemd.CheckUnitHasEnabledUFS(dbusConn, t.Unit)
		ufsStatus.Merge(status)
		status = ufsStatus
		err = helpers.MultiErrorAppend(e, err)
	}
	return status, err
}

// Apply tells the dbus to enable the unit
func (t *Enable) Apply() error {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	_, _, err = dbusConn.EnableUnitFiles([]string{t.Unit}, t.Runtime, t.Force)
	return err
}

func (t *Enable) Validate() error {
	if t.Unit == "" {
		return fmt.Errorf("task requires a %q parameter", "unit")
	}
	return nil
}
