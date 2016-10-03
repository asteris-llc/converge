---
title: "user.group"
slug: "user-group"
date: "2016-09-28T10:46:11-05:00"
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

- `gid` (uint32)

  Gid is the group gid.

- `name` (required string)

  Name is the group name.

- `state` (string)


  Valid values: `present` and `absent`

  State is whether the group should be present.


