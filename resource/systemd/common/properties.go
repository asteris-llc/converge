package common

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/helpers"
	"github.com/coreos/go-systemd/dbus"
	godbus "github.com/godbus/dbus"
)

// CheckProperty checks if the value of a unit matches one of the provided properties
func CheckProperty(conn *dbus.Conn, unit string, propertyName string, wants []dbus.Property) (status string, willChange bool, err error) {
	want, wants := wants[0], wants[1:]
	possibilities := want.Value.String()
	for _, w := range wants {
		possibilities = possibilities + ", " + w.Value.String()
	}
	prop, err := conn.GetUnitProperty(unit, propertyName)
	if err != nil {
		return err.Error(), false, err
	}
	if strings.Contains(possibilities, prop.Value.String()) {
		return fmt.Sprintf("property %q of unit %q is %q, expected one of %q", prop.Name, unit, prop.Value, possibilities), false, nil
	} else {
		return fmt.Sprintf("property %q of unit %q is %q, expected one of %q", prop.Name, unit, prop.Value, possibilities), true, nil
	}
}

// CheckProperties checks if every the actual values of a units properties matches those in wants.
func CheckProperties(conn *dbus.Conn, unit string, wants []dbus.Property) (status string, willChange bool, err error) {
	for _, want := range wants {
		prop, err := conn.GetUnitProperty(unit, want.Name)
		if err != nil {
			return err.Error(), false, err
		}

		if prop.Value.String() != want.Value.String() {
			status, willChange, err = helpers.SquashCheck(
				status, willChange, err,
				fmt.Sprintf("property %q of unit %q is %q, expected %q", prop.Name, unit, prop.Value, want.Value), true, nil)
		} else {
			status, willChange, err = helpers.SquashCheck(
				status, willChange, err,
				fmt.Sprintf("property %q of unit %q is %q, expected %q", prop.Name, unit, prop.Value, want.Value), false, nil)
		}
	}
	return status, willChange, err
}

func propDependency(name string, units []string) dbus.Property {
	return dbus.Property{
		Name:  name,
		Value: godbus.MakeVariant(units),
	}
}

type LoadState string

const (
	LSLoaded LoadState = "loaded"
	LSError  LoadState = "error"
	LSMasked LoadState = "masked"
)

func PropLoadState(ls LoadState) dbus.Property {
	return dbus.Property{
		Name:  "LoadState",
		Value: godbus.MakeVariant(string(ls)),
	}
}

type ActiveState string
type ActiveStates []ActiveState

const (
	ASActive       ActiveState = "active"
	ASReloading    ActiveState = "reloading"
	ASInactive     ActiveState = "inactive"
	ASFailed       ActiveState = "failed"
	ASActivating   ActiveState = "activating"
	ASDeactivating ActiveState = "deactivating"
)

func (states ActiveStates) Contains(state ActiveState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func PropActiveState(as ActiveState) dbus.Property {
	return dbus.Property{
		Name:  "ActiveState",
		Value: godbus.MakeVariant(string(as)),
	}
}

type UnitFileState string

const (
	UFSEnabled        UnitFileState = "enabled"
	UFSEnabledRuntime UnitFileState = "enabled-runtime"
	UFSLinked         UnitFileState = "linked"
	UFSLinkedRuntime  UnitFileState = "linked-runtime"
	UFSMasked         UnitFileState = "masked"
	UFSMaskedRuntime  UnitFileState = "masked-runtime"
	UFSStatic         UnitFileState = "static"
	UFSDisabled       UnitFileState = "disabled"
	UFSInvalid        UnitFileState = "invalid"
)

func PropUnitFileState(ufs UnitFileState) dbus.Property {
	return dbus.Property{
		Name:  "UnitFileState",
		Value: godbus.MakeVariant(string(ufs)),
	}
}
