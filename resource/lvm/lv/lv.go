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
	_, ok := r.lvs[r.name]
	r.needCreate = !ok

	ts := &resource.Status{}

	if r.needCreate {
		ts.Level = resource.StatusWillChange
	}
	return ts, nil
}

func (r *ResourceLV) Apply() (status resource.TaskStatus, err error) {
	if r.needCreate {
		if err := r.lvm.CreateLogicalVolume(r.group, r.name, r.size, r.sizeOption, r.sizeUnit); err != nil {
			return nil, err
		}
	}
	ts := &resource.Status{}
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
