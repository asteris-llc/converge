---
title: "user.user"
slug: "user-user"
date: "2016-09-16T15:26:27-05:00"
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

- `uid` (string)

  UID is the user ID.

- `gid` (string)

  Gid is the primary group ID for user.

- `username` (string)

  Username is the user login name.

- `name` (string)

  Name is the user description.

- `home_dir` (string)

  HomeDir is the user's login directory. By default,  the login
name is appended to the home directory.

- `state` (string)

  State is whether the user should be present.


