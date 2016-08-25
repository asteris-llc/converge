---
title: "module"
slug: "module"
date: "2016-08-24T23:41:00-05:00"
menu:
  main:
    parent: resources
---


Module remotely sources other modules and adds them to the tree


## Example

```hcl
param "message" {
  default = "Hello from another module!"
}

module "basic.hcl" "basic" {
  params = {
    message = "{{param `message`}}"
  }
}

```


## Parameters

- `params` (map of string to anything)

  Params is a map of strings to anything you'd like. It will be passed to
the called module as the default values for the `param`s there.


