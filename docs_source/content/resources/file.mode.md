---
title: "file.mode"
slug: "file-mode"
date: "2016-09-08T23:18:02-07:00"
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

- `mode` (octal string)

  Mode is the mode of the file, specified in octal.


