---
title: "package.rpm"
slug: "package-rpm"
date: "2016-12-16T11:20:35-06:00"
menu:
  main:
    parent: resources
---


RPM Package manages system packages with `rpm` and `yum`. It assumes that
both `rpm` and `yum` are installed on the system, and that the user has
permissions to install, remove, and query packages.


## Example

```hcl
package.rpm "mc" {
  name  = "mc"
  state = "present"
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `name` (required string)

  Name of the package or package group.

- `state` (State)


	Valid values: `present` and `absent`

  State of the package. Present means the package will be installed if
missing; Absent means the package will be uninstalled if present.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `name` (string)
  name of the package
 
- `state` (State)
  package state; one of "present" or "absent"
  

