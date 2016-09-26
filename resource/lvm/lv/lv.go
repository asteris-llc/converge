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
	lvs        map[string]*lowlevel.LogicalVolume
	needCreate bool
}

func (r *ResourceLV) Check(resource.Renderer) (status resource.TaskStatus, err error) {
	if _, ok := r.lvs[r.name]; !ok {
		r.needCreate = true
	} else {
		r.needCreate = false
	}

	ts := &resource.Status{
		Status:     r.name,
		WillChange: r.needCreate,
	}
	return ts, nil
}

func (r *ResourceLV) Apply(resource.Renderer) (status resource.TaskStatus, err error) {
	if r.needCreate {
		if err := r.lvm.CreateLogicalVolume(r.group, r.name, r.size, r.sizeOption, r.sizeUnit); err != nil {
			return nil, err
		}
	}
	ts := &resource.Status{
		Status:     r.name,
		WillChange: r.needCreate,
	}
	return ts, nil
}

func (r *ResourceLV) Setup(lvm lowlevel.LVM, group string, name string, sizeToParse string) error {
	r.group = group
	r.name = name
	r.lvm = lvm

	var err error
	r.lvs, err = r.lvm.QueryLogicalVolumes(r.group)
	if err != nil {
		return err
	}

	r.size, r.sizeOption, r.sizeUnit, err = lowlevel.ParseSize(sizeToParse)
	return err
}

func (r *ResourceLV) devicePath() string {
	return fmt.Sprintf("/dev/mapper/%s", r.name)
}

func init() {
	registry.Register("lvm.lv", (*Preparer)(nil), (*ResourceLV)(nil))
}
