package lv

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

// Preparer for LVM VG resource
type Preparer struct {
	Group string `hcl:"group",required:"true"`
	Name  string `hcl:"name",required:"true"`
	Size  string `hcl:"size",required:"true"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	size, err := lowlevel.ParseSize(p.Size)
	if err != nil {
		return nil, err
	}

	r := NewResourceLV(lowlevel.MakeLvmBackend(), p.Group, p.Name, size)
	return r, nil
}
