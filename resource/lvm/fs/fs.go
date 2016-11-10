package fs

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/load/registry"
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

// FIXME: RequiredBy statement should issued only when non-empty
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

	{
		fs, err := r.checkBlkid(r.mount.What)
		if err != nil {
			return nil, err
		}
		log.Debugf("blkid detect following fstype: %s, planned fstype: %s", fs, r.mount.Type)
		if fs == r.mount.Type {
			r.needMkfs = false
		} else if fs == "" {
			r.needMkfs = true
			status.AddDifference("format", fs, r.mount.Type, "")
		} else {
			return nil, fmt.Errorf("%s already contain other filesystem with different type %s", r.mount.What, fs)
		}
	}

	{
		ok, err := r.lvm.CheckUnit(r.unitFileName, r.unitFileContent)
		if err != nil {
			return nil, err
		}
		if ok {
			r.unitNeedUpdate = true
			status.AddDifference(r.unitFileName, "<none>", r.unitFileContent, "")
		}
	}

	// FIXME: check what device mounted to Where
	{
		ok, err := r.lvm.Mountpoint(r.mount.Where)
		if err != nil {
			return nil, err
		}
		r.mountNeedUpdate = r.unitNeedUpdate && !ok
	}

	if r.mountNeedUpdate {
		status.AddDifference(r.mount.Where, "<none>", fmt.Sprintf("mount %s", r.mount.Where), "")
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

	// FIXME: need mkdir
	if r.mountNeedUpdate {
		if err := r.lvm.StartUnit(r.unitServiceName()); err != nil {
			return nil, err
		}
	}

	return &resource.Status{}, nil
}

// FIXME: ugly kludge
func (r *resourceFS) checkBlkid(name string) (string, error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return "", nil
	}
	return r.lvm.Blkid(name)
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

func (r *resourceFS) escapedMountpoint() string {
	// FIXME: proper systemd' escaping of r.mountpoint should be
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

func init() {
	registry.Register("lvm.fs", (*Preparer)(nil), (*resourceFS)(nil))
}
