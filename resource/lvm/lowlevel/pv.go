package lowlevel

import (
	"strings"
)

type PV struct {
	Name   string `mapstructure:"LVM2_PV_NAME"`
	Vg     string `mapstructure:"LVM2_VG_NAME"`
	Device string
}

func QueryPV() (map[string]*PV, error) {
	result := map[string]*PV{}
	pvs, err := queryLVM("pvs", "pv_all,vg_name", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range pvs {
		pv := &PV{}
		if err := parseLVM(&pv, values); err != nil {
			return nil, err
		}
		if strings.HasPrefix(pv.Name, "/dev/dm-") {
			pv.Device, err = queryDeviceMapperName(pv.Name)
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
