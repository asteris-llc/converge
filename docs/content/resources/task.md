---
title: "task"
slug: "task"
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

Shell is a structure representing a task.

- CmdGenerator (`CommandExecutor`)

  
- CheckStmt (`string`)

  
- ApplyStmt (`string`)

  
- Status (`&{911 CommandResults}`)

  
- HealthStatus (`&{941 0xc42009c700}`)

  

