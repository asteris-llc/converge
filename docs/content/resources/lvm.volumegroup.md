---
title: "lvm.volumegroup"
slug: "lvm-volumegroup"
date: "2016-12-16T11:20:34-06:00"
menu:
  main:
    parent: resources
---


Volume group is responsible for creation LVM Volume Groups
from given block devices.


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


- `name` (required string)

  Name of created volume group

- `devices` (list of strings)

  Devices is list of entities to include into volume group

- `remove` (bool)

  Remove is enable removal devices omitted from `Devices` list from
from volume group

- `forceRemove` (bool)

  ForceRemove control destruction of volumes after removing
from volume group


