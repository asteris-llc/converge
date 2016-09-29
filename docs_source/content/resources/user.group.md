---
title: "user.group"
slug: "user-group"
date: "2016-09-21T14:04:24-05:00"
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

- `gid` (string)

  Gid is the group gid.

- `name` (string)

  Name is the group name.

- `state` (string)

  State is whether the group should be present.
Options are present and absent; default is present.


