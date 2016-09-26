package fs

import (
	"bytes"
	"fmt"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"strings"
	"text/template"
)

type ResourceFS struct {
	mount           *Mount
	lvm             lowlevel.LVM
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

func (r *ResourceFS) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{Status: r.mount.What}

	if fs, err := r.lvm.Blkid(r.mount.What); err != nil {
		return nil, err
	} else {
		if fs == r.mount.Type {
			r.needMkfs = false
		} else if fs == "" {
			r.needMkfs = true
			status.AddDifference("format", fs, r.mount.Type, "")
		} else {
			return nil, fmt.Errorf("%s already contain other filesystem with different type %s", r.mount.What, fs)
		}
	}

	if ok, err := r.lvm.CheckUnit(r.unitFileName, r.unitFileContent); err != nil {
		return nil, err
	} else if ok {
		r.unitNeedUpdate = true
		status.AddDifference(r.unitFileName, "<none>", r.unitFileContent, "")
	}

	// FIXME: check what device mounted to Where
	if ok, err := r.lvm.Mountpoint(r.mount.Where); err != nil {
		return nil, err
	} else {
		r.mountNeedUpdate = r.unitNeedUpdate && !ok
	}
	if r.mountNeedUpdate {
		status.AddDifference(r.mount.Where, "<none>", fmt.Sprintf("mount %s", r.mount.Where), "")
	}
	if resource.AnyChanges(status.Differences) {
		status.WillChange = true
		status.WarningLevel = resource.StatusWillChange
	}

	return status, nil
}

func (r *ResourceFS) Apply(resource.Renderer) (resource.TaskStatus, error) {
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

	// FIXME: need mkdir
	if r.mountNeedUpdate {
		// FIXME: abstraction leak
		if err := r.lvm.GetBackend().Run("systemctl", []string{"start", r.unitServiceName()}); err != nil {
			return nil, err
		}
	}

	return &resource.Status{
		Status: r.mount.What,
	}, nil
}

func (r *ResourceFS) Setup(lvm lowlevel.LVM, m *Mount) error {
	var err error
	r.lvm = lvm
	r.mount = m
	r.unitFileName = r.unitName()
	r.unitFileContent, err = r.renderUnitFile()
	if err != nil {
		return err
	}
	return nil
}

func (r *ResourceFS) escapedMountpoint() string {
	// FIXME: proper systemd' escaping of r.mountpoint should be
	return strings.Replace(strings.Trim(r.mount.Where, "/"), "/", "-", -1)
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

func init() {
	registry.Register("lvm.fs", (*Preparer)(nil), (*ResourceFS)(nil))
}
