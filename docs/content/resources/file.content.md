---
title: "file.content"
slug: "file-content"
date: "2016-12-16T11:20:34-06:00"
menu:
  main:
    parent: resources
---


Content renders content to disk


## Example

```hcl
param "message" {
  default = "Hello, World in {{param `filename`}}"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `content` (string)

  Content is the file content. This will be rendered as a template.

- `destination` (required string)

  Destination is the location on disk where the content will be rendered.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `content` (string)
  configured content of the file
 
- `destination` (string)
  configured destination of the file
  

