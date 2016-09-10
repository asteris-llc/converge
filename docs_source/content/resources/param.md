---
title: "param"
slug: "param"
date: "2016-09-08T23:18:03-07:00"
menu:
  main:
    parent: resources
---


Param controls the flow of values through `module` calls. You can use the
`{{param "name"}}` template call anywhere you need the value of a param
inside the current module.


## Example

```hcl
param "message" {
  default = "Hello, World!"
}

param "filename" {
  default = "test.txt"
}

task "render" {
  check = "cat {{param `filename`}} | tee /dev/stderr | grep -q '{{param `message`}}'"
  apply = "echo '{{param `message`}}' > {{param `filename`}}"
}

```


## Parameters

- `default` (optional string)

  Default is an optional field that provides a default value if none is
provided to this parameter. If this field is not set, this param will be
treated as required.


