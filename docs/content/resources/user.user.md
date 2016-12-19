---
title: "user.user"
slug: "user-user"
date: "2016-12-19T14:58:58-06:00"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `username` (required string)

  Username is the user login name.

- `new_username` (string)

  NewUsername is used when modifying a user.
Username will be changed to NewUsername. No changes to the home directory
name or location of the contents will be made. This can be done using
HomeDir and MoveDir options.

- `uid` (optional uint32)

  UID is the user ID.

- `groupname` (string)


	Only one of `gid` or `groupname` may be set.

  GroupName is the primary group for user and must already exist.
Only one of GID or Groupname may be indicated.

- `gid` (optional uint32)


	Only one of `gid` or `groupname` may be set.

  Gid is the primary group ID for user and must refer to an existing group.
Only one of GID or Groupname may be indicated.

- `name` (string)

  Name is the user description.
This field can be indicated when adding or modifying a user.

- `create_home` (bool)

  CreateHome when set to true will create the home directory for the user.
The files and directories contained in the skeleton directory (which can be
defined with the SkelDir option) will be copied to the home directory.

- `skel_dir` (string)

  SkelDir contains files and directories to be copied in the user's home
directory when adding a user. If not set, the skeleton directory is defined
by the SKEL variable in /etc/default/useradd or, by default, /etc/skel.
SkelDir is only valid is CreatHome is specified.

- `home_dir` (string)

  HomeDir is the name of the user's login directory. If not set, the home
directory is defined by appending the value of Username to the HOME
variable in /etc/default/useradd, resulting in /HOME/Username.
This field can be indicated when adding or modifying a user.

- `move_dir` (bool)

  MoveDir is used to move the contents of HomeDir when modifying a user.
HomeDir must also be indicated if MoveDir is set to true.

- `state` (State)


	Valid values: `present` and `absent`

  State is whether the user should be present.
The default value is present.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `username` (string)

  the configured username
 
- `newusername` (string)

  the desired username
 
- `uid` (string)

  the user id
 
- `groupname` (string)

  the group name
 
- `gid` (string)

  the group id
 
- `name` (string)

  the real name of the user
 
- `createhome` (bool)

  if the home directory should be created
 
- `skeldir` (string)

  the path to the skeleton directory
 
- `homedir` (string)

  the path to the home directory
 
- `movedir` (bool)

  if the contents of the home directory should be moved
 
- `state` (State)

  configured the user state
  

