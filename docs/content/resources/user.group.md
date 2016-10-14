---
title: "user.group"
slug: "user-group"
date: "2016-10-04T13:01:50-05:00"
menu:
  main:
    parent: resources
---


Group renders group data


## Example

```hcl
# create a group, only works on linux
user.group "group" {
  name = "test"
}

```


## Parameters

- `gid` (optional uint32)

  Gid is the group gid.

- `name` (required string)

  Name is the group name.

- `new_name` (string)

  NewName is used when modifying a group.
The group Name will be changed to NewName.

- `state` (State)


  Valid values: `present` and `absent`

  State is whether the group should be present.


