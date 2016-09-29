package systemd

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/coreos/go-systemd/dbus"
)

func CheckDaemonReload(conn *dbus.Conn, unit string) (bool, error) {
	prop, err := conn.GetUnitProperty(unit, "NeedDaemonReload")
	if err != nil {
		return false, err
	}
	shouldReload, ok := prop.Value.Value().(bool)
	return shouldReload && ok, nil
}

func ApplyDaemonReload() error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	err = conn.Connection.Reload()
	return err
}

const daemon_wont_reload_msg = "daemon does not need to be reloaded"

func GetDaemonReloadStatus(shouldReload bool) *resource.Status {
	status := daemon_wont_reload_msg
	warningLevel := resource.StatusNoChange
	if shouldReload {
		status = "daemon will be reloaded"
		warningLevel = resource.StatusWillChange
	}
	diffs := map[string]resource.Diff{
		"daemon-reload": resource.TextDiff{Default: daemon_wont_reload_msg, Values: [2]string{daemon_wont_reload_msg, status}},
	}
	return &resource.Status{
		Level:       warningLevel,
		Differences: diffs,
		Output:      []string{status},
	}
}
