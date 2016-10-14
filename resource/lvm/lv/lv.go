package lv

import (
	"fmt"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/wait"
)

type ResourceLV struct {
	group      string
	name       string
	size       int64
	sizeOption string
	sizeUnit   string
	lvm        lowlevel.LVM
	needCreate bool
	devicePath string
}

type Status struct {
	resource.Status
	DevicePath string
}

func (r *ResourceLV) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &Status{}
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
		status.Output = append(status.Output, fmt.Sprintf("group %s not exist, assume that it will be created"))
	}

	status.DevicePath = fmt.Sprintf("/dev/mapper/%s-%s", r.group, r.name)
	if r.needCreate {
		status.Level = resource.StatusWillChange
		status.AddDifference(fmt.Sprintf("%s", r.name), "<not exists>", fmt.Sprintf("created %s", status.DevicePath), "")
	}

	return status, nil
}

func (r *ResourceLV) Apply() (resource.TaskStatus, error) {
	status := &Status{}
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

	if devpath, err := r.deviceMapperPath(); err != nil {
		return nil, err
	} else {
		if devpath != r.devicePath {
			// FIXME: better put it to Messages, to log, both, upgrade to error???
			status.Output = append(status.Output, fmt.Sprintf("WARN: real device path '%s' diverge with planned '%s'", devpath, r.devicePath))
		}
		status.DevicePath = devpath
		if err := r.waitForDevice(devpath); err != nil {
			return status, err
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

func (r *ResourceLV) deviceMapperPath() (string, error) {
	lvs, err := r.lvm.QueryLogicalVolumes(r.group)
	if err != nil {
		return "", err
	}
	if lv, ok := lvs[r.name]; ok {
		return lv.DevicePath, nil
	}
	return "", fmt.Errorf("Can't find device mapper path for volume %s/%s", r.group, r.name)
}

func (r *ResourceLV) waitForDevice(path string) error {
	retrier := wait.PrepareRetrier("", "", 0)
	ok, err := retrier.RetryUntil(func() (bool, error) {
		// FIXME: Abstraction leak. Move all waitForDevice to LVM object?
		return r.lvm.GetBackend().Exists(path)
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("device path %s not appeared after %s seconds", path, retrier.Duration.String())
	}
	return nil
}

func init() {
	registry.Register("lvm.lv", (*Preparer)(nil), (*ResourceLV)(nil))
}
