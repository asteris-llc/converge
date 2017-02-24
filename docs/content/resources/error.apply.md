---
title: "error.apply"
slug: "error-apply"
date: "2017-02-24T09:23:17-06:00"
menu:
  main:
    parent: resources
---


Generates a runtime error in the graph


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

 

