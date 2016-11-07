---
title: "file.directory"
slug: "file-directory"
date: "2016-11-07T13:35:31-06:00"
menu:
  main:
    parent: resources
---


Directory makes sure a directory is present on disk


## Example

```hcl
# file.directory will create or ensure a directory is present
file.directory "hello" {
  destination = "hello"
}

# it can also accept a "create_all" parameter to do the equivalent of "mkdir -p"
file.directory "deeper" {
  destination = "deeper/a/b/c"
  create_all  = true
}

```


## Parameters

- `destination` (string)

  the location on disk to make the directory

- `create_all` (bool)

  whether or not to create all parent directories on the way up


