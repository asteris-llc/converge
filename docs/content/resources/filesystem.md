---
title: "filesystem"
slug: "filesystem"
date: "2016-12-08T15:04:22-06:00"
menu:
  main:
    parent: resources
---


Filesystem do formatting and mounting for LVM volumes
(also capable to format usual block devices as well)


## Example

```hcl
param "device" {
  default = "/dev/loop0"
}

lvm.volumegroup "vg-test" {
  name    = "test"
  devices = ["{{ param `device` }}"]
}

lvm.logicalvolume "lv-test" {
  group   = "test"
  name    = "test"
  size    = "1G"
  depends = ["lvm.volumegroup.vg-test"]
}

filesystem "mnt-me" {
  device  = "/dev/mapper/test-test"
  mount   = "/mnt"
  fstype  = "ext3"
  depends = ["lvm.logicalvolume.lv-test"]
}

```


## Parameters

- `device` (required string)

  Device path to be mount
Examples: `/dev/sda1`, `/dev/mapper/vg0-data`


- `mount` (required string)

  Mountpoint where device will be mounted
(should be an existing directory)
Example: /mnt/data


- `fstype` (required string)

  Fstype is filesystem type
(actually any linux filesystem, except `ZFS`)
Example:  `ext4`, `xfs`


- `requiredBy` (list of strings)

  RequiredBy is a list of dependencies, to pass to systemd .mount unit


- `wantedBy` (list of strings)

  WantedBy is a list of dependencies, to pass to systemd .mount unit


- `before` (list of strings)

  Before is a list of dependencies, to pass to systemd .mount unit




