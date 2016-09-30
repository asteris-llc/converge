---
title: "lvm.logicalvolume"
slug: "lvm-logicalvolume"
date: "2016-12-16T11:20:34-06:00"
menu:
  main:
    parent: resources
---


Logical volume creation


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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `group` (required string)

  Group where volume will be created

- `name` (required string)

  Name of volume, which will be created

- `size` (required string)

  Size of volume. Can be relative or absolute.
Relative size set in forms like `100%FREE`
(words after percent sign can be `FREE`, `VG`, `PV`)
Absolute size specified with suffix `BbKkMmGgTtPp`, upper case
suffix mean S.I. sizes (power of 10), lower case mean powers of 1024.
Also special suffixes `Ss`, which mean sectors.
Refer to LVM manpages for details.


