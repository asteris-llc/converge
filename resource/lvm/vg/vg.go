package vg

import (
	"fmt"
	"github.com/asteris-llc/converge/resource"
	lvm "github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type ResourceVG struct {
	Name            string
	Exists          bool
	DevicesToAdd    []string
	DevicesToRemove []string
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
        Status: "",
    }, nil
}

func (r *ResourceVG) Apply() error {
	if r.Exists {
		for _, d := range r.DevicesToAdd {
			if err := lvm.VGExtend(r.Name, d); err != nil {
				return err
			}
		}
		for _, d := range r.DevicesToRemove {
			if err := lvm.VGReduce(r.Name, d); err != nil {
				return err
			}
		}
	} else {
		return lvm.VGCreate(r.Name, r.DevicesToAdd)
	}
	return nil
}

func (r *ResourceVG) Setup(devs []string) error {
	pvs, err := lvm.QueryPV()
	if err != nil {
		return err
	}

	vgs, err := lvm.QueryVG()
	if err != nil {
		return err
	}

	for _, dev := range devs {
		if pv, ok := pvs[dev]; ok {
			if pv.Vg != r.Name {
				return fmt.Errorf("Can't add device %s to VG %s, it already member of VG %s", dev, r.Name, pv.Vg)
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
