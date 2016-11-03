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

package fs

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/pkg/errors"
)

type resourceFS struct {
	mount           *Mount
	lvm             lowlevel.LVM
	unitFileName    string
	unitFileContent string
	unitNeedUpdate  bool
	mountNeedUpdate bool
	needMkfs        bool
}

// Mount is a structure for holding values to be rendered as .mount unit for systemd
type Mount struct {
	What       string
	Where      string
	Type       string
	Before     string
	WantedBy   string
	RequiredBy string
}

// NB: RequiredBy statement should issued only when non-empty
// Related issue: https://github.com/asteris-llc/converge/issues/452
const unitTemplate = `[Unit]
Before=local-fs.target {{.Before}}

[Mount]
What={{.What}}
Where={{.Where}}
Type={{.Type}}

[Install]
WantedBy=local-fs.target {{.WantedBy}}
RequiredBy={{.RequiredBy}}`

func (r *resourceFS) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{}

	if err := r.lvm.CheckFilesystemTools(r.mount.Type); err != nil {
		return nil, errors.Wrap(err, "lvm.fs")
	}

	if err := r.checkBlkid(status); err != nil {
		return nil, err
	}

	if err := r.checkUnit(status); err != nil {
		return nil, err
	}

	if err := r.checkMountpoint(status); err != nil {
		return nil, err
	}

	if resource.AnyChanges(status.Differences) {
		status.Level = resource.StatusWillChange
	}

	return status, nil
}

func (r *resourceFS) Apply() (resource.TaskStatus, error) {
	if r.needMkfs {
		if err := r.lvm.Mkfs(r.mount.What, r.mount.Type); err != nil {
			return nil, err
		}
	}
	if r.unitNeedUpdate {
		if err := r.lvm.UpdateUnit(r.unitFileName, r.unitFileContent); err != nil {
			return nil, err
		}
	}

	if r.mountNeedUpdate {
		if err := r.lvm.StartUnit(r.unitServiceName()); err != nil {
			return nil, err
		}
	}

	return &resource.Status{}, nil
}

// NewResourceFS create new resource.Task node for create/mount FileSystem.
func NewResourceFS(lvm lowlevel.LVM, m *Mount) (resource.Task, error) {
	var err error
	r := &resourceFS{
		lvm:   lvm,
		mount: m,
	}
	r.unitFileName = r.unitName()
	r.unitFileContent, err = r.renderUnitFile()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *resourceFS) checkBlkid(status *resource.Status) error {
	fs, err := r.lvm.Blkid(r.mount.What)
	if err != nil {
		return errors.Wrapf(err, "retrieving current FS type of %s", r.mount.What)
	}
	log.Debugf("blkid detect following fstype: %s, planned fstype: %s", fs, r.mount.Type)
	if fs == r.mount.Type {
		r.needMkfs = false
	} else if fs == "" {
		r.needMkfs = true
		status.AddDifference("format", "<unformatted>", r.mount.Type, "")
	} else {
		return fmt.Errorf("%s already contain other filesystem with different type %s", r.mount.What, fs)
	}
	return nil
}

// NB: Here we need to ensure, that r. mount.Where is exists, and have proper permissions
// Related issue: https://github.com/asteris-llc/converge/issues/449
//
// NB: We need to ensure, that no other filesystems mounted to Where
// (now we check only if it mountpoint or not)
// Related issue: https://github.com/asteris-llc/converge/issues/450
func (r *resourceFS) checkMountpoint(status *resource.Status) error {
	ok, err := r.lvm.Mountpoint(r.mount.Where)
	if err != nil {
		return err
	}
	r.mountNeedUpdate = r.unitNeedUpdate && !ok

	if r.mountNeedUpdate {
		status.AddDifference(r.mount.Where, "<none>", fmt.Sprintf("mount %s", r.mount.Where), "")
	}
	return nil
}

func (r *resourceFS) checkUnit(status *resource.Status) error {
	ok, err := r.lvm.CheckUnit(r.unitFileName, r.unitFileContent)
	if err != nil {
		return err
	}
	if ok {
		r.unitNeedUpdate = true
		status.AddDifference(r.unitFileName, "<none>", r.unitFileContent, "")
	}
	return nil
}

func (r *resourceFS) escapedMountpoint() string {
	// NB: proper systemd' escaping of r.mountpoint should be implemented here
	// Related issue: https://github.com/asteris-llc/converge/issues/451
	return strings.Replace(strings.Trim(r.mount.Where, "/"), "/", "-", -1)
}

func (r *resourceFS) unitServiceName() string {
	return fmt.Sprintf("%s.mount", r.escapedMountpoint())
}

func (r *resourceFS) unitName() string {
	return fmt.Sprintf("/etc/systemd/system/%s", r.unitServiceName())
}

func (r *resourceFS) renderUnitFile() (string, error) {
	var b bytes.Buffer
	tmpl, err := template.New("unit.mount").Parse(unitTemplate)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&b, r.mount)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
