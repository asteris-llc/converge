---
title: "error"
slug: "error"
date: "2017-02-24T09:23:17-06:00"
menu:
  main:
    parent: resources
---


`error`, along with `error.plan` and `error.apply` provide a mechanism for
specifying runtime errors due to invalid conditions in an hcl file.
`error.plan` is an alias for `error`, and generates a runtime error that will
occur during the plan phase. `error.apply` will only raise an error during
application. Error resources will not cause a failure if they are in an
unevaluated branch of a conditional, and are a safe way of aborting execution
if a prerequisite requirement isn't met.


## Example

```hcl
switch "linux-only" {
  case "{{eq `linux` (platform.OS)}}" "linux" {
    task.query "ok" {
      query = "echo 'OK!'"
    }
  }

  default {
    error "not-linux" {
      error = "this module is only supported on Linux systems"
    }
  }
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `error` (required string)



## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.

 

