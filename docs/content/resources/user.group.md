---
title: "user.group"
slug: "user-group"
date: "2016-12-15T15:18:19-06:00"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


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
The default value is present.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `gid` (string)
  the configured group ID
 
- `name` (string)
  the configured group name
 
- `newname` (string)
  the desired group name
 
- `state` (State)
  the group state
  

