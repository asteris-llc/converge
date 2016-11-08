---
title: "package.apt"
slug: "package-apt"
date: "2016-11-01"
menu:
  main:
    parent: resources
---


apt Package manages system packages with `apt` and `dpkg`. It assumes that
both `apt` and `dpkg` are installed on the system, and that the user has
permissions to install, remove, and query packages.

## Example

```hcl
package.apt "mc" {
  name  = "mc"
  state = "present"
}

```

## Parameters

- `name` (required string)

  Name of the package or package group. Cannot be an empty string.

- `state` (State)


  Valid values: `present` and `absent`

  State of the package. Present means the package will be installed if
missing; Absent means the package will be uninstalled if present.


## Notes

Please note that only one `apt` command can run at a time. This may cause
errors as converge will run as many tasks in parallel as possible.

In order to get around the single-process limitation of the `apt` tool, you can:

* use a [`group`]({{< ref "dependencies.md#grouping" >}}) parameter for `package.apt` tasks.

* Pass multiple packages into the `name` field.

Below are examples of both techniques:

```hcl
package.apt "mc-vim" {
  name  = "mc vim"
  state = "present"
  group = "apt"
}

package.apt "git" {
  name  = "git"
  state = "present"
  group = "apt"
}


```