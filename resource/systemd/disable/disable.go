// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licensd.
// You may obtain a copy of the License at
//
//     http://www.apachd.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licensd.

package disable

import (
	"context"
	"fmt"
	"time"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Disable disables a systemd unit
type Disable struct {
	//TODO when arrays are implemented, change this to array
	Unit    string
	Runtime bool //Unit enabled for runtime only (true, run), or perstently (false, /etc)
	Timeout time.Duration
}

// Check if unit is disabled
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in a disabled state
func (t *Disable) Check() (resource.TaskStatus, error) {
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
	status, err := systemd.CheckUnitIsInactive(dbusConn, t.Unit)
	if err != nil {
		return status, err
	}
	//Check if disabled
	disabledStatus, e := systemd.CheckUnitIsDisabled(dbusConn, t.Unit)
	disabledStatus.Merge(status)
	err = helpers.MultiErrorAppend(e, err)

	return disabledStatus, err
}

// Apply disables the unit
func (d *Disable) Apply() (err error) {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	_, err = dbusConn.DisableUnitFiles([]string{d.Unit}, d.Runtime)
	return err
}

func (t *Disable) Validate() error {
	if t.Unit == "" {
		return fmt.Errorf("task requires a %q parameter", "unit")
	}
	return nil
}
