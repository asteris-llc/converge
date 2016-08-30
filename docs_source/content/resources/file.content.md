---
title: "file.content"
slug: "file-content"
date: "2016-08-30T15:12:06-05:00"
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

- `content` (string)

  Content is the file content. This will be rendered as a template.

- `destination` (string)

  Destination is the location on disk where the content will be rendered.


