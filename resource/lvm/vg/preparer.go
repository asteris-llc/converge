package vg

import (
	"path/filepath"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

// Preparer for LVM's Volume Group
type Preparer struct {
	Name    string   `hcl:"name",required:"true"`
	Devices []string `hcl:"devices"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// Device paths need to be real devices, not symlinks
	// (otherwise it breaks on GCE)
	devices := make([]string, len(p.Devices))
	for i, dev := range p.Devices {
		var err error
		devices[i], err = filepath.EvalSymlinks(dev)
		if err != nil {
			return nil, err
		}
	}

	rvg := NewResourceVG(lowlevel.MakeLvmBackend(), p.Name, devices)
	return rvg, nil
}
