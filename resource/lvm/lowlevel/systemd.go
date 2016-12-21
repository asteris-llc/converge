// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lowlevel

import "os"

func (lvm *realLVM) CheckUnit(filename string, content string) (bool, error) {
	realContent, err := lvm.backend.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	shouldUpdate := string(realContent) != content

	return shouldUpdate, nil
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
