---
title: "file.mode"
slug: "file-mode"
date: "2016-08-24T16:26:34-05:00"
menu:
  main:
    parent: modules
---

```hcl
param "filename" {
  default = "test.txt"
}

file.mode "render" {
  destination = "{{param `filename`}}"
  mode        = 0777
}
```

Mode monitors the file Mode of a file

- Destination (`string`)

  
- Mode (`&{os FileMode}`)

  

