---
title: "param"
slug: "param"
date: "2016-08-24T16:45:03-05:00"
menu:
  main:
    parent: resources
---

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

Param controls parameter flow inside execution

- Value (`string`)

  

