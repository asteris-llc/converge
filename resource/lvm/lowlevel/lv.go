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

// LogicalVolume is parsed record for LVM Logical Volume (from `lvs` output)
// Add more fields, if required
type LogicalVolume struct {
	Name       string `mapstructure:"LVM2_LV_NAME"`
	DevicePath string `mapstructure:"LVM2_LV_DM_PATH"`
}

func (lvm *realLVM) QueryLogicalVolumes(vg string) (map[string]*LogicalVolume, error) {
	result := map[string]*LogicalVolume{}
	lvs, err := lvm.Query("lvs", "all", []string{vg})
	if err != nil {
		return nil, err
	}
	for _, values := range lvs {
		lv := &LogicalVolume{}
		if err = lvm.parse(values, lv); err != nil {
			return nil, err
		}
		result[lv.Name] = lv
	}
	return result, nil
}
