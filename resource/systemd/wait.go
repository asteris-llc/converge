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

package systemd

import (
	"context"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/coreos/go-systemd/dbus"
)

const DefaultTimeout = time.Second * 5

func WaitToLoad(ctx context.Context, conn *dbus.Conn, unit string) error {
	//Check if unit is loading
	status, err := CheckProperty(conn, unit, "ActiveState", []*dbus.Property{
		PropActiveState(ASActivating),
		PropActiveState(ASReloading),
	})
	if err != nil {
		return err
	}
	//If loading wait until it becomes stable
	if status.WarningLevel == resource.StatusNoChange {
		err := WaitForActiveState(ctx, conn, unit, []ActiveState{ASActive, ASFailed, ASInactive})
		if err != nil {
			return err
		}
	}
	return nil
}

func WaitForActiveState(ctx context.Context, conn *dbus.Conn, unit string, states []ActiveState) error {
	err := conn.Subscribe()
	if err != nil {
		return err
	}

	statuses, errs := conn.SubscribeUnits(time.Second)
	// defer con.UnSubscribe()

	for {
		select {
		case status := <-statuses:
			unitStatus := status[unit]
			if unitStatus != nil {
				if ActiveStates(states).Contains(ActiveState(unitStatus.ActiveState)) {
					return nil
				}
			}
		case err := <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
