---
name: "file.mode"
slug: "file-mode"
date: "2016-08-24T13:20:32-05:00"
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

  

