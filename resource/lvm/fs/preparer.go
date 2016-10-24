package fs

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"

	"strings"
)

// Preparer for LVM FS Task
type Preparer struct {
	Device     string   `hcl:"device",required:"true"`
	Mount      string   `hcl:"mount",required:"true"`
	Fstype     string   `hcl:"fstype",reqired:"true"`
	RequiredBy []string `hcl:"requiredBy"`
	WantedBy   []string `hcl:"requiredBy"`
	Before     []string `hcl:"requiredBy"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	m := &Mount{
		What:       p.Device,
		Where:      p.Mount,
		Type:       p.Fstype,
		RequiredBy: strings.Join(p.RequiredBy, " "),
		WantedBy:   strings.Join(p.WantedBy, " "),
		Before:     strings.Join(p.Before, " "),
	}

	return NewResourceFS(lowlevel.MakeLvmBackend(), m)
}
