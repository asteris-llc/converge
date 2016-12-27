---
title: "file.owner"
slug: "file-owner"
date: "2016-12-22T11:43:15-06:00"
menu:
  main:
    parent: resources
---


Owner sets the file and group ownership of a file or directory.  If
`recursive` is set to true and `destination` is a directory, then it will
also recursively change ownership of all files and subdirectories.  Symlinks
are ignored.  If the file or directory does not exist during the plan phase
of application the differences will be calculated during application.
Otherwise changes will be limited to the files identified during the plan
phase of application.


## Example

```hcl
/* This file demonstrates proper usage of the file.owner module by creating a
 new file, then changing the ownership of that file to a different group. */

file.content "to-change" {
  destination = "tochange"
}

task.query "existing-group" {
  interpreter = "/bin/bash"
  query       = "echo -n $(ls -la {{lookup `file.content.to-change.destination`}} | awk '{print $4}')"
}

task.query "new-group" {
  interpreter = "/bin/bash"
  query       = "echo -n $(groups | xargs -n 1 echo | grep -v $(whoami) | grep -v {{lookup `task.query.existing-group.status.stdout`}} | head -n1)"
}

file.owner "owner-test" {
  destination = "{{lookup `file.content.to-change.destination`}}"
  group       = "{{lookup `task.query.new-group.status.stdout`}}"
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `destination` (required string)

  Destination is the location on disk where the content will be rendered.

- `recursive` (bool)

  Recursive indicates whether ownership changes should be applied
recursively.  Symlinks are not followed.

- `user` (string)

  Username specifies user-owernship by user name

- `uid` (optional int)


	Only one of `user` or `uid` may be set.

  UID specifies user-ownership by UID

- `group` (string)


	Only one of `group` or `gid` may be set.

  Groupname specifies group-ownership by groupname

- `gid` (optional int)


	Only one of `group` or `gid` may be set.

  GID specifies group ownership by gid

- `verbose` (bool)

  Verbose specifies that when performing recursive updates, a diff should be
shown for each file to be changed


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `destination` (string)

  the path to the file that should change; or if `recursive` is set, the path
to the root of the filesystem to recursively change.
 
- `username` (string)

  the username of the user that should be given ownership of the file
 
- `uid` (string)

  the uid of the user that should be given ownership of the file
 
- `group` (string)

  the group name of the group that should be given ownership of the file
 
- `gid` (string)

  the gid of the group that should be given ownership of the file
 
- `recursive` (bool)

  if true, and `destination` is a directory, apply changes recursively
  

