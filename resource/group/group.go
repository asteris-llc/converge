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

package group

import (
	"fmt"
	"os/user"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// State type for Group
type State string

const (
	// StatePresent indicates the group should be present
	StatePresent State = "present"

	// StateAbsent indicates the group should be absent
	StateAbsent State = "absent"
)

// Group manages user groups
type Group struct {
	GID     string
	Name    string
	NewName string
	State   State
	system  SystemUtils
}

// ModGroupOptions are the options specified in the configuration to be used
// when modifying a group
type ModGroupOptions struct {
	GID     string
	NewName string
}

// SystemUtils provides system utilities for group
type SystemUtils interface {
	AddGroup(groupName, groupID string) error
	DelGroup(groupName string) error
	ModGroup(groupName string, options *ModGroupOptions) error
	LookupGroup(groupName string) (*user.Group, error)
	LookupGroupID(groupID string) (*user.Group, error)
}

// ErrUnsupported is used when a system is not supported
var ErrUnsupported = fmt.Errorf("group: not supported on this system")

// NewGroup constructs and returns a new Group
func NewGroup(system SystemUtils) *Group {
	return &Group{
		system: system,
	}
}

// Check if a user group exists
func (g *Group) Check(resource.Renderer) (resource.TaskStatus, error) {
	var (
		groupByGid     *user.Group
		gidErr         error
		groupByNewName *user.Group
		newNameErr     error
	)

	// lookup the group by name and lookup the group by gid
	// the lookups return ErrUnsupported if the system is not supported
	// LookupGroup returns user.UnknownGroupError if the group is not found
	// LookupGroupID returns user.UnknownGroupIdError if the gid is not found
	groupByName, nameErr := g.system.LookupGroup(g.Name)
	if g.GID != "" {
		groupByGid, gidErr = g.system.LookupGroupID(g.GID)
	}
	if g.NewName != "" {
		groupByNewName, newNameErr = g.system.LookupGroup(g.NewName)
	}

	status := &resource.Status{}

	if nameErr == ErrUnsupported {
		status.Level = resource.StatusFatal
		return status, ErrUnsupported
	}

	switch g.State {
	case StatePresent:
		switch {
		case g.GID == "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)

			switch {
			case g.NewName == "":
				switch {
				case nameNotFound:
					status.Level = resource.StatusWillChange
					status.Output = append(status.Output, "add group")
					status.AddDifference("group", string(StateAbsent), fmt.Sprintf("group %s", g.Name), "")
				case groupByName != nil:
					status.Output = append(status.Output, fmt.Sprintf("group add: group %s already exists", g.Name))
				}
			case g.NewName != "":
				_, newNameNotFound := newNameErr.(user.UnknownGroupError)

				switch {
				case nameNotFound:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group modify: group %s does not exist", g.Name))
					return status, errors.New("cannot modify group")
				case newNameNotFound:
					status.Level = resource.StatusWillChange
					status.Output = append(status.Output, "modify group name")
					status.AddDifference("group", fmt.Sprintf("group %s", g.Name), fmt.Sprintf("group %s", g.NewName), "")
				case groupByNewName != nil:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group modify: group %s already exists", g.NewName))
					return status, errors.New("cannot modify group")
				}
			}
		case g.GID != "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)
			_, gidNotFound := gidErr.(user.UnknownGroupIdError)

			switch {
			case g.NewName == "":
				switch {
				case nameNotFound && gidNotFound:
					status.Level = resource.StatusWillChange
					status.Output = append(status.Output, "add group with gid")
					status.AddDifference("group", string(StateAbsent), fmt.Sprintf("group %s with gid %s", g.Name, g.GID), "")
				case nameNotFound:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group add: gid %s already exists", g.GID))
					return status, errors.New("cannot add group")
				case gidNotFound:
					status.Level = resource.StatusWillChange
					status.Output = append(status.Output, "modify group gid")
					status.AddDifference("group", fmt.Sprintf("group %s with gid %s", g.Name, groupByName.Gid), fmt.Sprintf("group %s with gid %s", g.Name, g.GID), "")
				case groupByName != nil && groupByGid != nil && groupByName.Name != groupByGid.Name || groupByName.Gid != groupByGid.Gid:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group add/modify: group %s and gid %s belong to different groups", g.Name, g.GID))
					return status, errors.New("cannot add or modify group")
				case groupByName != nil && groupByGid != nil && *groupByName == *groupByGid:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group add/modify: group %s with gid %s already exists", g.Name, g.GID))
					return status, errors.New("cannot add or modify group")
				}
			case g.NewName != "":
				_, newNameNotFound := newNameErr.(user.UnknownGroupError)

				switch {
				case newNameNotFound && gidNotFound:
					status.Level = resource.StatusWillChange
					status.Output = append(status.Output, "modify group name and gid")
					status.AddDifference("group", fmt.Sprintf("group %s with gid %s", g.Name, groupByName.Gid), fmt.Sprintf("group %s with gid %s", g.NewName, g.GID), "")
				case gidNotFound:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group modify: group %s already exists", g.NewName))
					return status, errors.New("cannot modify group")
				case newNameNotFound:
					status.Level = resource.StatusCantChange
					status.Output = append(status.Output, fmt.Sprintf("group modify: gid %s already exists", g.GID))
					return status, errors.New("cannot modify group")
				}
			}
		}
	case StateAbsent:
		switch {
		case g.GID == "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)

			switch {
			case nameNotFound:
				status.Output = append(status.Output, fmt.Sprintf("group delete: group %s does not exist", g.Name))
			case groupByName != nil:
				status.Level = resource.StatusWillChange
				status.Output = append(status.Output, "delete group")
				status.AddDifference("group", fmt.Sprintf("group %s", g.Name), string(StateAbsent), "")
			}
		case g.GID != "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)
			_, gidNotFound := gidErr.(user.UnknownGroupIdError)

			switch {
			case nameNotFound && gidNotFound:
				status.Output = append(status.Output, fmt.Sprintf("group delete: group %s and gid %s do not exist", g.Name, g.GID))
			case nameNotFound:
				status.Level = resource.StatusCantChange
				status.Output = append(status.Output, fmt.Sprintf("group delete: group %s does not exist", g.Name))
				return status, errors.New("cannot delete group")
			case gidNotFound:
				status.Level = resource.StatusCantChange
				status.Output = append(status.Output, fmt.Sprintf("group delete: gid %s does not exist", g.GID))
				return status, errors.New("cannot delete group")
			case groupByName != nil && groupByGid != nil && groupByName.Name != groupByGid.Name || groupByName.Gid != groupByGid.Gid:
				status.Level = resource.StatusCantChange
				status.Output = append(status.Output, fmt.Sprintf("group delete: group %s and gid %s belong to different groups", g.Name, g.GID))
				return status, errors.New("cannot delete group")
			case groupByName != nil && groupByGid != nil && *groupByName == *groupByGid:
				status.Level = resource.StatusWillChange
				status.Output = append(status.Output, "delete group with gid")
				status.AddDifference("group", fmt.Sprintf("group %s with gid %s", g.Name, g.GID), string(StateAbsent), "")
			}
		}
	default:
		status.Level = resource.StatusFatal
		return status, fmt.Errorf("group: unrecognized state %s", g.State)
	}

	return status, nil
}

// Apply changes for group
func (g *Group) Apply() (resource.TaskStatus, error) {
	var (
		groupByGid *user.Group
		gidErr     error
		newNameErr error
	)

	// lookup the group by name and lookup the group by gid
	// the lookups return ErrUnsupported if the system is not supported
	// LookupGroup returns user.UnknownGroupError if the group is not found
	// LookupGroupID returns user.UnknownGroupIdError if the gid is not found
	groupByName, nameErr := g.system.LookupGroup(g.Name)
	if g.GID != "" {
		groupByGid, gidErr = g.system.LookupGroupID(g.GID)
	}
	if g.NewName != "" {
		_, newNameErr = g.system.LookupGroup(g.NewName)
	}

	status := &resource.Status{}

	if nameErr == ErrUnsupported {
		status.Level = resource.StatusFatal
		return status, ErrUnsupported
	}

	switch g.State {
	case StatePresent:
		switch {
		case g.GID == "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)

			switch {
			case g.NewName == "":
				switch {
				case nameNotFound:
					err := g.system.AddGroup(g.Name, g.GID)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("error adding group %s", g.Name))
						return status, errors.Wrap(err, "group add")
					}
					status.Output = append(status.Output, fmt.Sprintf("added group %s", g.Name))
				default:
					status.Level = resource.StatusCantChange
					return status, fmt.Errorf("will not attempt add: group %s", g.Name)
				}
			case g.NewName != "":
				_, newNameNotFound := newNameErr.(user.UnknownGroupError)

				switch {
				case groupByName != nil && newNameNotFound:
					options := SetModGroupOptions(g)
					err := g.system.ModGroup(g.Name, options)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("error modifying group %s", g.Name))
						return status, errors.Wrap(err, "group modify")
					}
					status.Output = append(status.Output, fmt.Sprintf("modified group %s with new name %s", g.Name, g.NewName))
				default:
					status.Level = resource.StatusCantChange
					return status, fmt.Errorf("will not attempt modify: group %s", g.Name)
				}
			}
		case g.GID != "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)
			_, gidNotFound := gidErr.(user.UnknownGroupIdError)

			switch {
			case g.NewName == "":
				switch {
				case nameNotFound && gidNotFound:
					err := g.system.AddGroup(g.Name, g.GID)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("error adding group %s with gid %s", g.Name, g.GID))
						return status, errors.Wrap(err, "group add")
					}
					status.Output = append(status.Output, fmt.Sprintf("added group %s with gid %s", g.Name, g.GID))
				case gidNotFound:
					options := SetModGroupOptions(g)
					err := g.system.ModGroup(g.Name, options)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("error modifying group %s with new gid %s", g.Name, g.GID))
						return status, errors.Wrap(err, "group modify")
					}
					status.Output = append(status.Output, fmt.Sprintf("modified group %s with new gid %s", g.Name, g.GID))
				default:
					status.Level = resource.StatusCantChange
					return status, fmt.Errorf("will not attempt add/modify: group %s with gid %s", g.Name, g.GID)
				}
			case g.NewName != "":
				_, newNameNotFound := newNameErr.(user.UnknownGroupError)

				switch {
				case groupByName != nil && newNameNotFound && gidNotFound:
					options := SetModGroupOptions(g)
					err := g.system.ModGroup(g.Name, options)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("error modifying group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID))
						return status, errors.Wrap(err, "group modify")
					}
					status.Output = append(status.Output, fmt.Sprintf("modified group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID))
				default:
					status.Level = resource.StatusCantChange
					return status, fmt.Errorf("will not attempt modify: group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID)
				}
			}
		}
	case StateAbsent:
		switch {
		case g.GID == "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)

			switch {
			case !nameNotFound && groupByName != nil:
				err := g.system.DelGroup(g.Name)
				if err != nil {
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error deleting group %s", g.Name))
					return status, errors.Wrap(err, "group delete")
				}
				status.Output = append(status.Output, fmt.Sprintf("deleted group %s", g.Name))
			default:
				status.Level = resource.StatusCantChange
				return status, fmt.Errorf("will not attempt delete: group %s", g.Name)
			}
		case g.GID != "":
			_, nameNotFound := nameErr.(user.UnknownGroupError)
			_, gidNotFound := gidErr.(user.UnknownGroupIdError)

			switch {
			case !nameNotFound && !gidNotFound && groupByName != nil && groupByGid != nil && *groupByName == *groupByGid:
				err := g.system.DelGroup(g.Name)
				if err != nil {
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error deleting group %s with gid %s", g.Name, g.GID))
					return status, errors.Wrap(err, "group delete")
				}
				status.Output = append(status.Output, fmt.Sprintf("deleted group %s with gid %s", g.Name, g.GID))
			default:
				status.Level = resource.StatusCantChange
				return status, fmt.Errorf("will not attempt delete: group %s with gid %s", g.Name, g.GID)
			}
		}
	default:
		status.Level = resource.StatusFatal
		return status, fmt.Errorf("group: unrecognized state %s", g.State)
	}

	return status, nil
}

// SetModGroupOptions returns a ModGroupOptions struct with the options
// specified in the configuration for modifying a group
func SetModGroupOptions(g *Group) *ModGroupOptions {
	options := new(ModGroupOptions)

	if g.GID != "" {
		options.GID = g.GID
	}

	if g.NewName != "" {
		options.NewName = g.NewName
	}

	return options
}
