---
title: "File"
slug: "file"
date: "2016-10-04"
menu:
  main:
    parent: resources
---

Manages files, file content, directories, hard and soft links.

## Example

```hcl
param "dir" {
  default = "/tmp/converge"
}

file "dir" {
  destination = "{{param `dir`}}"
  type        = "directory"
  state       = "present"
}

file "hello" {
  destination = "{{param `dir`}}/hello.txt"
  mode        = "0750"
  depends     = ["file.dir"]
}

file "symlink" {
  destination = "{{param `dir`}}/symlink"
  target      = "{{lookup `file.hello.Destination`}}"
  state       = "present"
  type        = "symlink"
}

file "deepdir" {
  destination = "{{param `dir`}}/a/b/c/d/e"
  type        = "directory"
  state       = "present"
}

# create parent dir using force
file "deepfile" {
  destination = "{{param `dir`}}/a/b/c/d/e/f/g/deep.txt"
  type        = "file"
  state       = "present"
  content     = "converge test"
  force       = "true"
}
```

## Parameters

- `destination` (required string)

  Destination specifies which file will be modified by this resource.

- `force` (boolean)

  Force the change. This is requried for:

  - changing the file type
  - changing an existing hard link
  - creating a file where the parent directory does not exist

  Default is `false`

- `group` (string)

  The group account that owns the file. This account must exist on the system or an
  error will be raised.

  Default is the effective Group ID of the converge process. If unset, group ownership
  of existing files will not be changed.

- `mode` (required base 8 optional uint32)

  Mode is the mode of the file, specified in octal.

  Default is the `0750`. If unset, permissions on existin files will not be changed.

- `state` (string)

  Whether the file should be `present` or `absent`. If `absent` is requested,
  the file will be removed from the system.

  Default is `present`

- `target` (string)

  Required for hard and soft links, this is the target file that `destination`
  is linked to. `destination -> target`

  This file must exist for a hard link, or an error will be raised.

  Set `force = true` to change the target of a hard link.

- `user` (string)

  The user account that owns the file. This account must exist on the system or an
  error will be raised.

  Default is the effective User ID of the converge process. If unset, the ownership
  of existing files will not be changed.
