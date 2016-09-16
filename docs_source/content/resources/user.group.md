---
title: "user.group"
slug: "user-group"
date: "2016-09-16T15:32:02-05:00"
menu:
  main:
    parent: resources
---


Group renders group data


## Example

```hcl
# create a group, only works on linux
user.group "group" {
  gid   = "123"
  name  = "test"
  state = "present"
}

```


## Parameters

- `gid` (string)

  Gid is the group gid.

- `name` (string)

  Name is the group name.

- `state` (string)

  State is whether the group should be present.


