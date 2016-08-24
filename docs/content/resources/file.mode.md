---
title: "file.mode"
slug: "file-mode"
date: "2016-08-24T16:45:03-05:00"
menu:
  main:
    parent: resources
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

  

