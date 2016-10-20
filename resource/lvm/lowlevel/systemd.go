package lowlevel

import (
	"os"
)

func (lvm *realLVM) CheckUnit(filename string, content string) (bool, error) {
	realContent, err := lvm.backend.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return string(realContent) != content, nil
}

func (lvm *realLVM) UpdateUnit(filename string, content string) error {
	if err := lvm.backend.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	return lvm.backend.Run("systemctl", []string{"daemon-reload"})
}

func (lvm *realLVM) StartUnit(unitname string) error {
	return lvm.backend.Run("systemctl", []string{"start", unitname})
}
