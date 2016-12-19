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
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

// ErrInvalidStat is returned when we fail to stat a file
var ErrInvalidStat = errors.New("Underlying OS did not return a valid stat_t")

// OSExecutor provides a real implementation of OSProxy
type OSExecutor struct{}

// Walk returns the result of filepath.Walk
func (o *OSExecutor) Walk(root string, f filepath.WalkFunc) error {
	return filepath.Walk(root, f)
}

// Chown returns the result of OS.Chown
func (o *OSExecutor) Chown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

func osStat(path string) (*syscall.Stat_t, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	statT, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, ErrInvalidStat
	}
	if statT == nil {
		return nil, ErrInvalidStat
	}
	return statT, nil
}

// GetUID returns the UID of a file or directory
func (o *OSExecutor) GetUID(path string) (int, error) {
	statT, err := osStat(path)
	if err != nil {
		return 0, err
	}
	return int(statT.Uid), nil
}

// GetGID returns the GID of a file or directory
func (o *OSExecutor) GetGID(path string) (int, error) {
	statT, err := osStat(path)
	if err != nil {
		return 0, err
	}
	return int(statT.Gid), nil
}

// LookupGroupID proxies user.LookupGroupID
func (o *OSExecutor) LookupGroupID(s string) (*user.Group, error) {
	return user.LookupGroupId(s)
}

// LookupGroup proxies user.LookupGroup
func (o *OSExecutor) LookupGroup(s string) (*user.Group, error) {
	return user.LookupGroup(s)
}

// LookupID proxies user.LookupID
func (o *OSExecutor) LookupID(s string) (*user.User, error) {
	return user.LookupId(s)
}

// Lookup proxies user.Lookup
func (o *OSExecutor) Lookup(s string) (*user.User, error) {
	return user.Lookup(s)
}
