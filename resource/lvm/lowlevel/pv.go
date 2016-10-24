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
