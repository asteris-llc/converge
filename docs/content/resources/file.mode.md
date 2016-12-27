---
title: "file.mode"
slug: "file-mode"
date: "2016-12-22T11:43:14-06:00"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `destination` (required string)

  Destination specifies which file will be modified by this resource. The
file must exist on the system (for example, having been created with
`file.content`.)

- `mode` (required base 8 optional uint32)

  Mode is the mode of the file, specified in octal.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `destination` (string)

  path to the file that will be modified
 
- `mode` (os.FileMode)

  the mode that the file or directory should be configured with
  

