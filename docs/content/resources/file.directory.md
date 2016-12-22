---
title: "file.directory"
slug: "file-directory"
date: "2016-12-22T11:43:14-06:00"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `destination` (required string)

  the location on disk to make the directory

- `create_all` (bool)

  whether or not to create all parent directories on the way up


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `destination` (string)

  directory path
 
- `createall` (bool)

  if true, directories will be created recursively
  

