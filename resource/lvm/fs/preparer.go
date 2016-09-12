package fs

import (
	"github.com/asteris-llc/converge/resource"
)

type Preparer struct {
	Device string `hcl:"device"`
	Mount  string `hcl:"mount"`
	Fstype string `hcl:"fstype"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	device, err := render.Render("device", p.Device)
	if err != nil {
		return nil, err
	}
	mount, err := render.Render("mount", p.Mount)
	if err != nil {
		return nil, err
	}
	fstype, err := render.Render("fstype", p.Fstype)
	if err != nil {
		return nil, err
	}

	r := &ResourceFS{
		mount: &Mount{
			What:       device,
			Where:      mount,
			Type:       fstype,
			RequiredBy: "", // FIXME: render it
			WantedBy:   "", // FIXME: render it
			Before:     "", // FIXME: render it
		},
	}

	err = r.Setup()
	return r, err
}
