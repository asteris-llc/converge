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

import (
	"strings"
)

// PhysicalVolume is parsed record for LVM Physical Volume (from `pvs` output)
// Add more fields, if required
type PhysicalVolume struct {
	Name   string `mapstructure:"LVM2_PV_NAME"`
	Group  string `mapstructure:"LVM2_VG_NAME"`
	Device string
}

func (lvm *realLVM) QueryPhysicalVolumes() (map[string]*PhysicalVolume, error) {
	result := map[string]*PhysicalVolume{}
	pvs, err := lvm.Query("pvs", "pv_all,vg_name", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range pvs {
		pv := &PhysicalVolume{}
		if err := lvm.parse(values, pv); err != nil {
			return nil, err
		}
		if strings.HasPrefix(pv.Name, "/dev/dm-") {
			pv.Device, err = lvm.QueryDeviceMapperName(pv.Name)
			if err != nil {
				return nil, err
			}
		} else {
			pv.Device = pv.Name
		}
		result[pv.Device] = pv
	}
	return result, nil
}
