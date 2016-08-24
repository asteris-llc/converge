---
title: "file.content"
slug: "file-content"
date: "2016-08-24T16:55:17-05:00"
menu:
  main:
    parent: resources
---

Content renders content to disk

## Example
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

## Parameters
- Content (`string`)

  the file content. Will be rendered as a template.   

- Destination (`string`)

  the location on disk that the content will end up at   


