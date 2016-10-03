---
title: "platform"
slug: "platform"
date: "2016-10-03"
menu:
  main:
    parent: content
---


Platform retrieves read-only system information from the underlying host. Returns
an empty string if the value is not set.

## Example

```hcl
file.content "platformData" {
  destination = "{{platform.OS}}-information.txt"
  content     = "Detected {{platform.Name}} ({{platform.OS}}) {{platform.Version}} {{platform.LinuxDistribution}}"
}
```

## Parameters

- `Build` (string)
  The build number. On Linux this is the value of the LSB `BUILD_ID`.
  On macOS the value is the `BuildVersion` obtained from `/usr/bin/sw_vers`.
  Examples: `15G31` (macOS)

- `OS` (string)
  The underlying OS. This value is the golang [runtime.GOOS](https://golang.org/pkg/runtime/).
  Examples: `darwin` (Apple macOS), `linux`.

- `LinuxDistribution` (string)
  Value of the [LSB](https://www.freedesktop.org/software/systemd/man/os-release.html) `ID` in `/etc/os-release`.
  Examples: "alpine", "centos", "coreos", "debian", "nixos", "ubuntu".

- `LinuxLSBLike` (list of strings)
  Value of LSB `ID_LIKE` in `/etc/os-release` to identify related distributions.
  Examples: `Centos` distributions return `{"rhel", "fedora"}`.
  `Ubuntu` distributions have this value set to `{"debian"}`.

- `Name` (string)
  Value of LSB `NAME` in `/etc/os-release` for Linux, [`/usr/bin/sw_vers`](https://developer.apple.com/legacy/library/documentation/Darwin/Reference/ManPages/man1/sw_vers.1.html) `ProductName` on macOS.
  Operating System Name. Examples: `CoreOS`, `Debian`, `Mac OS X`, `NixOS`, `Ubuntu`

- `PrettyName` (string)
  Longer name of the operating system. Taken from the LSB value of `PRETTY_NAME`.
  Examples: `Alpine Linux v3.4`, `CoreOS 835.9.0`, `Debian GNU/Linux 8 (jessie)`, `Ubuntu 16.04.1 LTS`

- `Version` (string)
  The version of the operating system. `/usr/bin/sw_vers` `ProductVersion` on macOS.
  On Linux systems, this is the value of LSB `VERSION_ID`.
  Examples: `10.11.6` (macOS), `835.9.0` (coreOS), `8` (debian), `16.04` (ubuntu)
