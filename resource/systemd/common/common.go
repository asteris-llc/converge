package common

import (
	"fmt"
	"time"

	"github.com/coreos/go-systemd/dbus"
)

const DefaultTimeout = time.Second * 5

func WaitToLoad(conn *dbus.Conn, unit string, timeout time.Duration) error {
	//Check if unit is loading
	_, isLoading, _ := CheckProperty(conn, unit, "ActiveState", []dbus.Property{
		PropActiveState(ASActivating),
		PropActiveState(ASReloading),
	})
	//If loading wait until it becomes stable
	if isLoading {
		err := WaitForActiveState(conn, unit, []ActiveState{ASActive, ASFailed, ASInactive}, timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func WaitForActiveState(conn *dbus.Conn, unit string, states []ActiveState, timeout time.Duration) error {
	err := conn.Subscribe()
	if err != nil {
		return err
	}
	statuses, errs := conn.SubscribeUnits(time.Second * 1)

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
		case <-time.After(timeout):
			return fmt.Errorf("waiting for one of the following active states: %s timed out after %s", states, timeout)
		}
	}
}

func CheckUnitIsActive(conn *dbus.Conn, unit string) (status string, willChange bool, err error) {
	validActiveStates := []dbus.Property{
		PropActiveState(ASActive),
	}
	return CheckProperty(conn, unit, "ActiveState", validActiveStates)
}

func CheckUnitIsInactive(conn *dbus.Conn, unit string) (status string, willChange bool, err error) {
	validActiveStates := []dbus.Property{
		PropActiveState(ASInactive),
	}
	return CheckProperty(conn, unit, "ActiveState", validActiveStates)
}

func CheckUnitHasValidUFS(conn *dbus.Conn, unit string) (status string, willChange bool, err error) {
	validRuntimeStates := []dbus.Property{
		PropUnitFileState(UFSEnabled),
		PropUnitFileState(UFSLinked),
		PropUnitFileState(UFSMasked),
		PropUnitFileState(UFSStatic),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}

func CheckUnitHasValidUFSRuntimes(conn *dbus.Conn, unit string) (status string, willChange bool, err error) {
	validRuntimeStates := []dbus.Property{
		PropUnitFileState(UFSEnabledRuntime),
		PropUnitFileState(UFSLinkedRuntime),
		PropUnitFileState(UFSMaskedRuntime),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}

func CheckUnitIsDisabled(conn *dbus.Conn, unit string) (status string, willChange bool, err error) {
	validRuntimeStates := []dbus.Property{
		PropUnitFileState(UFSDisabled),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}
