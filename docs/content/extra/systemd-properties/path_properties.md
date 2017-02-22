---
title: "systemd_PathTypeProperties"
slug: "systemd_PathTypeProperties"
date: "2017-02-03T07:38:25-06:00"
menu:
  main:
    parent: resources
---
## PathTypeProperties

PathTypeProperties represent the properties that can be set for path units. The
fields and their exported types are listed below; for a more comprehensive
explanation of these fields see:
https://www.freedesktop.org/wiki/Software/systemd/dbus/


- `DirectoryMode` (`uint32`)
- `MakeDirectory` (`bool`)
- `Paths` (`[][]interface{}`)
- `Result` (`string`)
- `Unit` (`string`)

