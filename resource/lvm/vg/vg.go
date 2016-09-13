package vg

import (
	"fmt"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type ResourceVG struct {
	Name            string
	Exists          bool
	DevicesToAdd    []string
	DevicesToRemove []string
	lvm             *lowlevel.LVM
}

func (r *ResourceVG) Check() (status resource.TaskStatus, err error) {
	var wc bool
	if r.Exists && len(r.DevicesToAdd) == 0 && len(r.DevicesToRemove) == 0 {
		wc = false
	} else {
		wc = true
	}
	return &resource.Status{
		WillChange: wc,
		Status:     "",
	}, nil
}

func (r *ResourceVG) Apply() error {
	if r.Exists {
		for _, d := range r.DevicesToAdd {
			if err := r.lvm.ExtendVolumeGroup(r.Name, d); err != nil {
				return err
			}
		}
		for _, d := range r.DevicesToRemove {
			if err := r.lvm.ReduceVolumeGroup(r.Name, d); err != nil {
				return err
			}
		}
	} else {
		return r.lvm.CreateVolumeGroup(r.Name, r.DevicesToAdd)
	}
	return nil
}

func (r *ResourceVG) Setup(devs []string) error {
	r.lvm = lowlevel.MakeLvmBackend()
	pvs, err := r.lvm.QueryPhysicalVolumes()
	if err != nil {
		return err
	}

	vgs, err := r.lvm.QueryVolumeGroups()
	if err != nil {
		return err
	}

	for _, dev := range devs {
		if pv, ok := pvs[dev]; ok {
			if pv.Group != r.Name {
				return fmt.Errorf("Can't add device %s to VG %s, it already member of VG %s", dev, r.Name, pv.Group)
			}
		} else {
			r.DevicesToAdd = append(r.DevicesToAdd, dev)
		}
	}

	// FIXME: something better here? May be go able handle sets of strings?
	for d, _ := range pvs {
		found := false
		for _, d2 := range devs {
			if d2 == d {
				found = true
			}
		}
		if !found {
			r.DevicesToRemove = append(r.DevicesToRemove, d)
		}
	}

	if _, ok := vgs[r.Name]; ok {
		r.Exists = true
	}
	return nil
}

func init() {
	registry.Register("lvm.vg", (*Preparer)(nil), (*ResourceVG)(nil))
}
