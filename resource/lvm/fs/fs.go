package fs

import (
	"bytes"
	"fmt"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"io/ioutil"
	"strings"
	"text/template"
)

type ResourceFS struct {
	mount           *Mount
	lvm             *lowlevel.LVM
	unitFileName    string
	unitFileContent string
	unitNeedUpdate  bool
	mountNeedUpdate bool
	needMkfs        bool
}

type Mount struct {
	What       string
	Where      string
	Type       string
	Before     string
	WantedBy   string
	RequiredBy string
}

// FIXME: RequiredBy statement should issued only when non-empty
const UNIT_TEMPLATE = `[Unit]
Before=local-fs.target {{.Before}}

[Mount]
What={{.What}}
Where={{.Where}}
Type={{.Type}}

[Install]
WantedBy=local-fs.target {{.WantedBy}}
RequiredBy={{.RequiredBy}}`

func (r *ResourceFS) Check() (status resource.TaskStatus, err error) {
	if fs, err := r.lvm.Blkid(r.mount.What); err != nil {
		return nil, err
	} else {
		if fs == r.mount.Type {
			r.needMkfs = false
		} else if fs == "" {
			r.needMkfs = true
		} else {
			return nil, fmt.Errorf("%s already contain filesystem %s", r.mount.What, fs)
		}
	}

	if unit, err := ioutil.ReadFile(r.unitFileName); err != nil {
		return nil, err
	} else {
		r.unitNeedUpdate = string(unit) != r.unitFileContent
	}

	// FIXME: check what device mounted to Where
	if ok, err := r.lvm.Mountpoint(r.mount.Where); err != nil {
		return nil, err
	} else {
		r.mountNeedUpdate = r.unitNeedUpdate && !ok
	}

	return &resource.Status{
		WillChange: r.needMkfs && r.unitNeedUpdate && r.mountNeedUpdate,
		Status:     "",
	}, nil
}

func (r *ResourceFS) Apply() error {
	if r.needMkfs {
		if err := r.lvm.Mkfs(r.mount.What, r.mount.Type); err != nil {
			return err
		}
	}
	if r.unitNeedUpdate {
		if err := ioutil.WriteFile(r.unitFileName, []byte(r.unitFileContent), 0644); err != nil {
			return err
		}
		if err := r.lvm.Backend.Run("systemctl", []string{"daemon-reload"}); err != nil {
			return err
		}
	}
	if r.mountNeedUpdate {
		if err := r.lvm.Backend.Run("systemctl", []string{"start", r.unitServiceName()}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ResourceFS) Setup() error {
	var err error
	r.lvm = lowlevel.MakeLvmBackend()
	r.unitFileName = r.unitName()
	r.unitFileContent, err = r.renderUnitFile()
	if err != nil {
		return err
	}
	return nil
}

func (r *ResourceFS) escapedMountpoint() string {
	// FIXME: proper systemd' escaping of r.mountpoint should be
	return strings.Replace(r.mount.Where, "/", "-", -1)
}

func (r *ResourceFS) unitServiceName() string {
	return fmt.Sprintf("%s.mount", r.escapedMountpoint())
}

func (r *ResourceFS) unitName() string {
	return fmt.Sprintf("/etc/systemd/system/%s", r.unitServiceName())
}

func (r *ResourceFS) renderUnitFile() (string, error) {
	var b bytes.Buffer
	tmpl, err := template.New("unit.mount").Parse(UNIT_TEMPLATE)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&b, r.mount)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
