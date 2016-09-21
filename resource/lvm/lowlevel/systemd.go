package lowlevel

import (
	"os"
)

func (lvm *RealLVM) CheckUnit(filename string, content string) (bool, error) {
	realContent, err := lvm.Backend.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return string(realContent) != content, nil
}

func (lvm *RealLVM) UpdateUnit(filename string, content string) error {
	if err := lvm.Backend.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	return lvm.Backend.Run("systemctl", []string{"daemon-reload"})
}
