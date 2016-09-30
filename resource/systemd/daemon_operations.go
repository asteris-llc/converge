package systemd

// CheckDaemonReload checks whether the systemd daemon needs to be reloaded.
// it does not check if a unit is failed and needs to be reset though.
func CheckDaemonReload(unit string) (bool, error) {
	conn, err := GetDbusConnection()
	if err != nil {
		return false, err
	}
	prop, err := conn.Connection.GetUnitProperty(unit, "NeedDaemonReload")
	if err != nil {
		return false, err
	}
	shouldReload, ok := prop.Value.Value().(bool)
	return shouldReload && ok, nil
}

// ApplyDaemonReload reloads the daemon
func ApplyDaemonReload() error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	err = conn.Connection.Reload()
	return err
}

// CheckResetFailed checks whether the fail state of the unit should be
// reset
func CheckResetFailed(unit string) (bool, error) {
	conn, err := GetDbusConnection()
	if err != nil {
		return false, err
	}
	prop, err := conn.Connection.GetUnitProperty(unit, "ActiveState")
	if err != nil {
		return false, err
	}
	shouldReset, ok := prop.Value.Value().(string)
	return ok && ASFailed.Equal(ActiveState(shouldReset)), nil
}

// ApplyResetFailed resets the failed state of a unit.
func ApplyResetFailed(unit string) error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	err = conn.Connection.ResetFailedUnit(unit)
	return err
}

const daemonWontReloadMsg = "daemon does not need to be reloaded"
