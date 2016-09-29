package systemd

import "github.com/coreos/go-systemd/dbus"

//CheckDaemonReload checks whether the systemd daemon needs to be reloaded.
//it does not check if a unit is failed and needs to be reset though.
//NOTE : should I implement checkResetFailed?
func CheckDaemonReload(conn *dbus.Conn, unit string) (bool, error) {
	prop, err := conn.GetUnitProperty(unit, "NeedDaemonReload")
	if err != nil {
		return false, err
	}
	shouldReload, ok := prop.Value.Value().(bool)
	return shouldReload && ok, nil
}

//ApplyDaemonReload reloads the daemon
//NOTE : should I implement applyResetFailed?
func ApplyDaemonReload() error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	err = conn.Connection.Reload()
	return err
}

const daemonWontReloadMsg = "daemon does not need to be reloaded"
