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

// VolumeGroup is parsed record for LVM Volume Groups (from `vgs` output)
// Add more fields, if required
// (at the moment we need only LVM2_VG_NAME to get list all existing groups)
type VolumeGroup struct {
	Name string `mapstructure:"LVM2_VG_NAME"`
}

func (lvm *realLVM) QueryVolumeGroups() (map[string]*VolumeGroup, error) {
	result := map[string]*VolumeGroup{}
	vgs, err := lvm.Query("vgs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range vgs {
		vg := &VolumeGroup{}
		if err = lvm.parse(values, vg); err != nil {
			return nil, err
		}
		result[vg.Name] = vg
	}
	return result, nil
}
