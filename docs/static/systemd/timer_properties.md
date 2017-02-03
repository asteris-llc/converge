## TimerTypeProperties

TimerTypeProperties represent the properties that can be set for timer
units. The fields and their exported types are listed below; for a more
comprehensive explanation of these fields see:
https://www.freedesktop.org/wiki/Software/systemd/dbus/

- `AccuracyUSec` (uint64)
- `LastTriggerUSec` (uint64)
- `LastTriggerUSecMonotonic` (uint64)
- `NextElapseUSecMonotonic` (uint64)
- `NextElapseUSecRealtime` (uint64)
- `Persistent` (bool)
- `RandomizedDelayUSec` (uint64)
- `RemainAfterElapse` (bool)
- `Result` (string)
- `TimersCalendar` ([][]interface{})
- `TimersMonotonic` ([][]interface{})
- `Unit` (string)
- `WakeSystem` (bool)
