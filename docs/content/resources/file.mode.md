---
title: "file.mode"
slug: "file-mode"
date: "2016-08-24T16:55:17-05:00"
menu:
  main:
    parent: resources
---

Mode monitors the file Mode of a file

## Example
```hcl
param "filename" {
  default = "test.txt"
}

file.mode "render" {
  destination = "{{param `filename`}}"
  mode        = 0777
}
```

## Parameters
- Destination (`string`)

  
- Mode (`&{os FileMode}`)

  

