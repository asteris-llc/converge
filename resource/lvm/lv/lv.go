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
	lvm        *lowlevel.LVM
	lvs        map[string]*lowlevel.LogicalVolume
	needCreate bool
}

func (r *ResourceLV) Check() (status resource.TaskStatus, err error) {
	if _, ok := r.lvs[r.name]; !ok {
		r.needCreate = true
	} else {
		r.needCreate = false
	}

	ts := &resource.Status{
		Status:     "",
		WillChange: r.needCreate,
	}
	return ts, nil
}

func (r *ResourceLV) Apply() error {
	if r.needCreate {
		if err := r.lvm.CreateLogicalVolume(r.group, r.name, r.size, r.sizeOption, r.sizeUnit); err != nil {
			return err
		}
	}
	return nil
}

func (r *ResourceLV) Setup(sizeToParse string) error {
	var err error
	r.lvm = lowlevel.MakeLvmBackend()
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
