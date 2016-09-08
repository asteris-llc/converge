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
	"fmt"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/coreos/go-systemd/dbus"
)

const DefaultTimeout = time.Second * 5

func WaitToLoad(ctx context.Context, conn *dbus.Conn, unit string) error {
	// first checck if there was an error loading ocnfiguration
	loadStatus, err := CheckProperty(conn, unit, "LoadState", []*dbus.Property{
		PropLoadState(LSError),
		PropLoadState(LSNotFound),
	})
	if err != nil {
		return err
	}
	if loadStatus.WarningLevel == resource.StatusNoChange {
		loadError, err := conn.GetUnitProperty(unit, "LoadError")
		if err != nil {
			return fmt.Errorf("configuration failed to load")
		}
		return fmt.Errorf("%s: %q", loadError.Name, loadError.Value.String())
	}

	//Check if unit is activating
	status, err := CheckProperty(conn, unit, "ActiveState", []*dbus.Property{
		PropActiveState(ASReloading),
		PropActiveState(ASActivating),
		PropActiveState(ASDeactivating),
	})
	if err != nil {
		return err
	}
	//If loading wait until it becomes stable
	if status.WarningLevel == resource.StatusNoChange {
		err := WaitForLoadedState(ctx, conn, unit)
		if err != nil {
			return err
		}
	}
	return nil
}

func WaitForLoadedState(ctx context.Context, conn *dbus.Conn, unit string) error {
	err := conn.Subscribe()
	if err != nil {
		return err
	}
	// @ Reviewer. function may complete faster by putting this function in a
	// looping gorutine constatnly pooling for the state instead of subscribing to state changes
	// ever arbitrary time period. this would take more cpu cycles.
	statuses, errs := conn.SubscribeUnits(100 * time.Millisecond)
	for {
		select {
		case status := <-statuses:
			unitStatus, ok := status[unit]
			if unitStatus != nil && ok {
				if !ActiveState(unitStatus.ActiveState).Equal(ASReloading) {
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
