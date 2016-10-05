---
title: "file.mode"
slug: "file-mode"
date: "2016-10-04T13:01:49-05:00"
menu:
  main:
    parent: resources
---


Mode monitors the mode of a file


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

- `destination` (required string)

  Destination specifies which file will be modified by this resource. The
file must exist on the system (for example, having been created with
`file.content`.)

- `mode` (required base 8 optional uint32)

  Mode is the mode of the file, specified in octal.


