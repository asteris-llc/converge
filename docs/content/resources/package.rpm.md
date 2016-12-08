---
title: "package.rpm"
slug: "package-rpm"
date: "2016-12-08T15:04:23-06:00"
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

- `name` (required string)

  Name of the package or package group.


- `state` (State)


	Valid values: `present` and `absent`

  State of the package. Present means the package will be installed if
missing; Absent means the package will be uninstalled if present.



## Exported Fields
- `name` (string)
  name of the package
 
- `state` (State)
  package state; one of "present" or "absent"
  

