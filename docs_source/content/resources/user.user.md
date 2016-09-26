---
title: "user.user"
slug: "user-user"
date: "2016-09-28T10:46:11-05:00"
menu:
  main:
    parent: resources
---


User renders user data


## Example

```hcl
# create a user, only works on linux
user.user "user" {
  username = "test"
}

```


## Parameters

- `username` (required string)

  Username is the user login name.

- `uid` (uint32)

  UID is the user ID.

- `groupname` (string)


  Only one of `gid` or `groupname` may be set.

  GroupName is the primary group for user and must already exist.
Only one of GID or Groupname may be indicated.

- `gid` (uint32)


  Only one of `gid` or `groupname` may be set.

  Gid is the primary group ID for user and must refer to an existing group.
Only one of GID or Groupname may be indicated.

- `name` (string)

  Name is the user description.

- `home_dir` (string)

  HomeDir is the user's login directory. By default,  the login
name is appended to the home directory.

- `state` (string)


  Valid values: `present` and `absent`

  State is whether the user should be present.


