package vg

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type Preparer struct {
	Name    string   `hcl:"name"`
	Devices []string `hcl:"devices"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}

	devices := make([]string, len(p.Devices))
	for i, dev := range p.Devices {
		rdev, err := render.Render("devices["+string(i)+"]", dev)
		if err != nil {
			return nil, err
		}

		// also resolve symlink here
		devices[i], err = resolveSymlink(rdev)
		if err != nil {
			return nil, err
		}
	}

	rvg := &ResourceVG{
		Name: name,
	}

	rvg.Setup(lowlevel.MakeLvmBackend(), devices)

	return rvg, nil
}
