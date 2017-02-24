---
title: "systemd_TimerTypeProperties"
slug: "systemd_TimerTypeProperties"
date: "2017-02-03T07:38:25-06:00"
menu:
  main:
    parent: resources
---
## TimerTypeProperties

TimerTypeProperties represent the properties that can be set for timer
units. The fields and their exported types are listed below; for a more
comprehensive explanation of these fields see:
https://www.freedesktop.org/wiki/Software/systemd/dbus/

- `AccuracyUSec` (`uint64`)
- `LastTriggerUSec` (`uint64`)
- `LastTriggerUSecMonotonic` (`uint64`)
- `NextElapseUSecMonotonic` (`uint64`)
- `NextElapseUSecRealtime` (`uint64`)
- `Persistent` (`bool`)
- `RandomizedDelayUSec` (`uint64`)
- `RemainAfterElapse` (`bool`)
- `Result` (`string`)
- `TimersCalendar` (`[][]interface{}`)
- `TimersMonotonic` (`[][]interface{}`)
- `Unit` (`string`)
- `WakeSystem` (`bool`)

