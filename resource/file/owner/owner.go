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

package owner

import (
	"os/user"
	"path/filepath"

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// OSProxy is an intermediary used for interfacing with the underlying OS or
// test mocks.
type OSProxy interface {
	Walk(string, filepath.WalkFunc) error
	Chown(string, int, int) error
	GetUID(string) (int, error)
	GetGID(string) (int, error)
	LookupGroupId(string) (*user.Group, error)
	LookupGroup(string) (*user.Group, error)
	LookupId(string) (*user.User, error)
	Lookup(string) (*user.User, error)
}

// Owner represents the ownership mode of a file or directory
type Owner struct {
	Destination string `export:"destination"`
	Username    string `export:"username"`
	UID         string `export:"uid"`
	Group       string `export:"group"`
	GID         string `export:"gid"`
	Recursive   bool   `export:"recursive"`
	executor    OSProxy
}

// SetOSProxy sets the internal OS Proxy
func (o *Owner) SetOSProxy(p OSProxy) *Owner {
	o.executor = p
	return o
}

// Check checks the ownership status of the file
func (o *Owner) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// Apply applies the ownership to the file
func (o *Owner) Apply(context.Context) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}
