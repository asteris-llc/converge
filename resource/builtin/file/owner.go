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

package file

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/asteris-llc/converge/resource"
)

// Owner controls the owner of the underlying resource
type Owner struct {
	resource.DependencyTracker `hcl:",squash"`

	Name           string
	RawDestination string `hcl:"destination"`
	RawOwner       string `hcl:"owner"`
	RawUID         string `hcl:"uid"`

	destination string
	owner       string
	uid         string
	gid         string
	renderer    *resource.Renderer
}

// Prepare this resource for use
func (o *Owner) Prepare(parent *resource.Module) (err error) {
	if o.RawOwner != "" && o.RawUID != "" {
		return &resource.ValidationError{
			Location: o.String() + ".owner",
			Err:      fmt.Errorf("A username or uid may be provided for file.owner. Not both"),
		}
	}

	o.renderer, err = resource.NewRenderer(parent)

	// render destination
	o.destination, err = o.renderer.Render(o.String()+".destination", o.RawDestination)
	if err != nil {
		return err
	}

	// render owner
	o.owner, err = o.renderer.Render(o.String()+".owner", o.RawOwner)
	if err != nil {
		return err
	}
	// render uid
	o.uid, err = o.renderer.Render(o.String()+".uid", o.RawUID)
	if err != nil {
		return err
	}
	// Validate ids
	var u *user.User
	if o.owner != "" {
		u, err = user.Lookup(o.owner)
	} else if o.uid != "" {
		u, err = user.LookupId(o.uid)
	} else {
		return &resource.ValidationError{
			Location: o.String() + ".owner",
			Err:      fmt.Errorf("Empty username field."),
		}
	}
	//Failed to find user.
	if err != nil {
		return err
	}

	o.owner = u.Username
	o.uid = u.Uid
	o.gid = u.Gid

	// render dependencies
	err = o.DependencyTracker.ComputeDependencies(
		o.String()+".dependencies",
		o.renderer,
		o.RawDestination,
		o.RawOwner,
	)

	if err != nil {
		return err
	}

	return nil
}

func (o *Owner) String() string {
	return "file.owner." + o.Name
}

// SetName modifies the name of this module
func (o *Owner) SetName(name string) {
	o.Name = name
}

// Check whether the destination has the right owner
func (o *Owner) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(o.destination)
	if err != nil {
		return
	}

	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		err = &resource.ValidationError{
			Location: o.String() + ".destination",
			Err:      fmt.Errorf("file.owner does not currently work on non linux systems"),
		}
		return
	}
	uid := statT.Uid
	u, err := user.LookupId(fmt.Sprintf("%v", uid))
	if err != nil {
		return
	}
	return u.Username, u.Username != o.owner, nil
}

// Apply changes the owner
func (o *Owner) Apply() (err error) {
	uid, _ := strconv.Atoi(o.uid)
	gid, _ := strconv.Atoi(o.gid)
	return os.Chown(o.destination, uid, gid)
}
