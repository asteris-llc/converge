---
title: "module"
slug: "module"
date: "2016-08-24T16:55:17-05:00"
menu:
  main:
    parent: resources
---

Module holds stringified values for parameters

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
- Params (`&{764 string string}`)

  

