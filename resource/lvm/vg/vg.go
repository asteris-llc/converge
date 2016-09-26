package vg

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type ResourceVG struct {
	Name            string
	Exists          bool
	DevicesToAdd    []string
	DevicesToRemove []string
	lvm             lowlevel.LVM
	DeviceList      []string
}

func (r *ResourceVG) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{Status: r.Name}

	pvs, err := r.lvm.QueryPhysicalVolumes()
	if err != nil {
		return nil, err
	}

	// check if group exists
	if vgs, err := r.lvm.QueryVolumeGroups(); err != nil {
		return nil, err
	} else {
		_, r.Exists = vgs[r.Name]
	}

	// process new devices
	for _, dev := range r.DeviceList {
		if pv, ok := pvs[dev]; ok {
			if pv.Group != r.Name {
				return nil, fmt.Errorf("Can't add device %s to VG %s, it already member of VG %s", dev, r.Name, pv.Group)
			}
		} else {
			r.DevicesToAdd = append(r.DevicesToAdd, dev)
			status.AddDifference(dev, "<none>", fmt.Sprintf("member of volume group %s", r.Name), "")
		}
	}

	// process removed devices
	for d, _ := range pvs {
		found := false
		for _, d2 := range r.DeviceList {
			if d2 == d {
				found = true
			}
		}
		if !found {
			r.DevicesToRemove = append(r.DevicesToRemove, d)
			status.AddDifference(d, fmt.Sprintf("member of volume group %s", r.Name), "<removed>", "")
		}
	}

	if !r.Exists {
		status.AddDifference(r.Name, "<not exists>", strings.Join(r.DevicesToAdd, ", "), "")
	}

	if resource.AnyChanges(status.Differences) {
		status.WillChange = true
		status.WarningLevel = resource.StatusWillChange
	}
	return status, nil
}

func (r *ResourceVG) Apply(resource.Renderer) (status resource.TaskStatus, err error) {
	if r.Exists {
		for _, d := range r.DevicesToAdd {
			if err := r.lvm.ExtendVolumeGroup(r.Name, d); err != nil {
				return nil, err
			}
		}
		for _, d := range r.DevicesToRemove {
			if err := r.lvm.ReduceVolumeGroup(r.Name, d); err != nil {
				return nil, err
			}
		}
	} else {
		if err := r.lvm.CreateVolumeGroup(r.Name, r.DevicesToAdd); err != nil {
			return nil, err
		}
	}

	return &resource.Status{
		Status: r.Name,
	}, nil
}

func (r *ResourceVG) Setup(lvm lowlevel.LVM, devs []string) {
	r.lvm = lvm
	r.DeviceList = devs
}

func init() {
	registry.Register("lvm.vg", (*Preparer)(nil), (*ResourceVG)(nil))
}
