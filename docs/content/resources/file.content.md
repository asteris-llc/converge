---
title: "file.content"
slug: "file-content"
date: "2016-08-24T23:23:56-05:00"
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

- Content (string)

  
- Destination (string)

  

