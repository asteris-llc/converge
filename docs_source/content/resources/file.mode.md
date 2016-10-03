---
title: "file.mode"
slug: "file-mode"
date: "2016-09-27T18:19:08-05:00"
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

- `destination` (string)

  Destination specifies which file will be modified by this resource. The
file must exist on the system (for example, having been created with
`file.content`.)

- `mode` (base 8 uint32)

  Mode is the mode of the file, specified in octal.


