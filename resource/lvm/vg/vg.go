package vg

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/pkg/errors"
)

type resourceVG struct {
	name       string
	deviceList []string
	lvm        lowlevel.LVM

	exists          bool
	devicesToAdd    []string
	devicesToRemove []string
}

func (r *resourceVG) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{}

	if err := r.lvm.Check(); err != nil {
		return nil, errors.Wrap(err, "lvm.vg")
	}

	pvs, err := r.lvm.QueryPhysicalVolumes()
	if err != nil {
		return nil, err
	}

	// check if group exists
	{
		vgs, err := r.lvm.QueryVolumeGroups()
		if err != nil {
			return nil, err
		}
		_, r.exists = vgs[r.name]
	}

	// process new devices
	for _, dev := range r.deviceList {
		if pv, ok := pvs[dev]; ok {
			if pv.Group != r.name {
				return nil, fmt.Errorf("Can't add device %s to VG %s, it already member of VG %s", dev, r.name, pv.Group)
			}
		} else {
			r.devicesToAdd = append(r.devicesToAdd, dev)
			status.AddDifference(dev, "<none>", fmt.Sprintf("member of volume group %s", r.name), "")
		}
	}

	// process removed devices
	for d, _ := range pvs {
		found := false
		for _, d2 := range r.deviceList {
			if d2 == d {
				found = true
			}
		}
		if !found {
			if pv, ok := pvs[d]; ok && pv.Group == r.name {
				r.devicesToRemove = append(r.devicesToRemove, d)
				status.AddDifference(d, fmt.Sprintf("member of volume group %s", r.name), "<removed>", "")
			}
		}
	}

	if !r.exists {
		status.AddDifference(r.name, "<not exists>", strings.Join(r.devicesToAdd, ", "), "")
	}

	if resource.AnyChanges(status.Differences) {
		status.Level = resource.StatusWillChange
	}
	return status, nil
}

func (r *resourceVG) Apply() (status resource.TaskStatus, err error) {
	if r.exists {
		for _, d := range r.devicesToAdd {
			if err := r.lvm.ExtendVolumeGroup(r.name, d); err != nil {
				return nil, err
			}
		}
		for _, d := range r.devicesToRemove {
			if err := r.lvm.ReduceVolumeGroup(r.name, d); err != nil {
				return nil, err
			}
		}
	} else {
		if err := r.lvm.CreateVolumeGroup(r.name, r.devicesToAdd); err != nil {
			return nil, err
		}
	}

	return &resource.Status{}, nil
}

// NewResourceVG creates new resource.Task node for Volume Group
func NewResourceVG(lvm lowlevel.LVM, name string, devs []string) resource.Task {
	return &resourceVG{
		lvm:        lvm,
		deviceList: devs,
		name:       name,
	}
}

func init() {
	registry.Register("lvm.vg", (*Preparer)(nil), (*resourceVG)(nil))
}
