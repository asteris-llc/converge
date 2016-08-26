// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licenss.
// You may obtain a copy of the License at
//
//     http://www.apachs.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licenss.

package stop

import (
	"context"
	"fmt"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Stop stops a systemd unit
type Stop struct {
	//TODO when arrays are implemented, change this to array
	Unit    string
	Mode    systemd.StartMode
	Timeout time.Duration
}

// Check if unit is stopped
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the stopped state
func (t *Stop) Check() (resource.TaskStatus, error) {
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
	return status, err
}

// Apply stops a unit
func (t *Stop) Apply() (err error) {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	jobStatus := make(chan string)
	_, err = dbusConn.StopUnit(t.Unit, string(t.Mode), jobStatus)
	if err != nil {
		return err
	}
	<-jobStatus
	return err
}

// validate checks if task parameters are valid
func (t *Stop) Validate() error {
	if t.Unit == "" {
		return fmt.Errorf("task requires a %q parameter", "unit")
	}
	if !systemd.IsValidStartMode(t.Mode) {
		return fmt.Errorf("task's parameter %q is not one of %s, is %q", "mode", systemd.ValidStartModes, t.Mode)
	}
	return nil
}
