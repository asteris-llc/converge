package lv

import (
	"fmt"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type ResourceLV struct {
	group      string
	name       string
	size       int64
	sizeOption string
	sizeUnit   string
	lvm        lowlevel.LVM
	needCreate bool
}

func (r *ResourceLV) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{}
	ok, err := r.checkVG()
	if err != nil {
		return nil, err
	}

	if ok {
		lvs, err := r.lvm.QueryLogicalVolumes(r.group)
		if err != nil {
			return nil, err
		}

		_, ok = lvs[r.name]
		r.needCreate = !ok
	} else {
		status.AddDifference(fmt.Sprintf("group: %s", r.group), "<not exists>", "created", "")
	}

	if r.needCreate {
		status.Level = resource.StatusWillChange
		status.AddDifference(fmt.Sprintf("volume: %s", r.name), "<not exists>", "created", "")
	}
	return status, nil
}

func (r *ResourceLV) Apply() (resource.TaskStatus, error) {
	status := &resource.Status{}
	if ok, err := r.checkVG(); err != nil {
		return nil, err
	} else {
		if !ok {
			return nil, fmt.Errorf("Group %s not exists", r.group)
		}
	}
	if r.needCreate {
		if err := r.lvm.CreateLogicalVolume(r.group, r.name, r.size, r.sizeOption, r.sizeUnit); err != nil {
			return nil, err
		}
	}
	return status, nil
}

func (r *ResourceLV) Setup(lvm lowlevel.LVM, group string, name string, sizeToParse string) error {
	r.group = group
	r.name = name
	r.lvm = lvm

	var err error
	r.size, r.sizeOption, r.sizeUnit, err = lowlevel.ParseSize(sizeToParse)
	return err
}

func (r *ResourceLV) checkVG() (bool, error) {
	vgs, err := r.lvm.QueryVolumeGroups()
	if err != nil {
		return false, err
	}
	_, ok := vgs[r.group]
	return ok, nil
}

func (r *ResourceLV) devicePath() string {
	return fmt.Sprintf("/dev/mapper/%s", r.name)
}

func init() {
	registry.Register("lvm.lv", (*Preparer)(nil), (*ResourceLV)(nil))
}
