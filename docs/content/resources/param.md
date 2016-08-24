---
title: "param"
slug: "param"
date: "2016-08-24T16:55:17-05:00"
menu:
  main:
    parent: resources
---

Param controls parameter flow inside execution

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
- Value (`string`)

  

