// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lv

import (
	"fmt"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/pkg/errors"
)

type resourceLV struct {
	group      string
	name       string
	size       *lowlevel.LvmSize
	lvm        lowlevel.LVM
	needCreate bool
	devicePath string
}

// Status is a resource.Status extended by DevicePath of created volume
type Status struct {
	resource.Status
	DevicePath string
}

func (r *resourceLV) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &Status{}

	// Check for LVM prerequizites
	if err := r.lvm.Check(); err != nil {
		return nil, errors.Wrap(err, "lvm.lv")
	}

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
		r.needCreate = true
	}

	status.DevicePath = fmt.Sprintf("/dev/mapper/%s-%s", r.group, r.name)
	if r.needCreate {
		status.Level = resource.StatusWillChange
		status.AddDifference(fmt.Sprintf("%s", r.name), "<not exists>", fmt.Sprintf("created %s", status.DevicePath), "")
	}

	return status, nil
}

func (r *resourceLV) Apply() (resource.TaskStatus, error) {
	status := &Status{}
	{
		ok, err := r.checkVG()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("Group %s not exists", r.group)
		}
	}

	if r.needCreate {
		if err := r.lvm.CreateLogicalVolume(r.group, r.name, r.size); err != nil {
			return nil, err
		}
	}

	{
		devpath, err := r.deviceMapperPath()
		if err != nil {
			return nil, err
		}
		if devpath != r.devicePath {
			// FIXME: better put it to Messages, to log, both, upgrade to error???
			status.Output = append(status.Output, fmt.Sprintf("WARN: real device path '%s' diverge with planned '%s'", devpath, r.devicePath))
		}
		status.DevicePath = devpath
		if err := r.lvm.WaitForDevice(devpath); err != nil {
			return status, err
		}
	}
	return status, nil
}

// NewResourceLV create new resource.Task node for LVM Volume Groups
func NewResourceLV(lvm lowlevel.LVM, group string, name string, size *lowlevel.LvmSize) resource.Task {
	return &resourceLV{
		group: group,
		name:  name,
		lvm:   lvm,
		size:  size,
	}
}

func (r *resourceLV) checkVG() (bool, error) {
	vgs, err := r.lvm.QueryVolumeGroups()
	if err != nil {
		return false, err
	}
	_, ok := vgs[r.group]
	return ok, nil
}

func (r *resourceLV) deviceMapperPath() (string, error) {
	lvs, err := r.lvm.QueryLogicalVolumes(r.group)
	if err != nil {
		return "", err
	}
	if lv, ok := lvs[r.name]; ok {
		return lv.DevicePath, nil
	}
	return "", fmt.Errorf("Can't find device mapper path for volume %s/%s", r.group, r.name)
}

func init() {
	registry.Register("lvm.lv", (*Preparer)(nil), (*resourceLV)(nil))
}
