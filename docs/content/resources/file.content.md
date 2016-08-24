---
title: "file.content"
slug: "file-content"
date: "2016-08-24T16:45:03-05:00"
menu:
  main:
    parent: resources
---

```hcl
param "message" {
  default = "Hello, World"
}

param "filename" {
  default = "test.txt"
}

file.content "render" {
  destination = "{{param `filename`}}"
  content     = "{{param `message`}}"
}
```

Content renders content to disk

- Content (`string`)

  the file content. Will be rendered as a template.   

- Destination (`string`)

  the location on disk that the content will end up at   


