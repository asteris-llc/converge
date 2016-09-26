package lv

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

type Preparer struct {
	Group string `hcl:"group"`
	Name  string `hcl:"name"`
	Size  string `hcl:"size"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	group, err := render.Render("group", p.Group)
	if err != nil {
		return nil, err
	}
	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}
	size, err := render.Render("size", p.Size)
	if err != nil {
		return nil, err
	}

	r := &ResourceLV{}
	err = r.Setup(lowlevel.MakeLvmBackend(), group, name, size)
	return r, err
}
