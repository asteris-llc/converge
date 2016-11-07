param "device" {
    default = "/dev/sdb"
}

lvm.volumegroup "vg-test" {
  name = "test"
  devices = [ "{{ param `device` }}" ]
}

lvm.logicalvolume "lv-test" {
  group = "test"
  name = "test"
  size = "1G"
  depends  = [ "lvm.volumegroup.vg-test" ]
}

filesystem "mnt-me"  {
  device = "/dev/mapper/test-test"
  mount = "/mnt"
  fstype = "xfs"
  depends = [ "lvm.logicalvolume.lv-test" ]
}
