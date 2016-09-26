---
title: "user.user"
slug: "user-user"
date: "2016-09-26T09:00:07-05:00"
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

- `username` (string)

  Username is the user login name.

- `uid` (string)

  UID is the user ID.

- `groupname` (string)

  Groupname is the primary group for user and must already exist.
Only one of GID or Groupname may be indicated.

- `gid` (string)

  Gid is the primary group ID for user and must refer to an existing group.
Only one of GID or Groupname may be indicated.

- `name` (string)

  Name is the user description.

- `home_dir` (string)

  HomeDir is the user's login directory. By default,  the login
name is appended to the home directory.

- `state` (string)

  State is whether the user should be present.
Options are present and absent; default is present.


