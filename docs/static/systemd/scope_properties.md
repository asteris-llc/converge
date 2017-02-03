## ScopeTypeProperties

ScopeTypeProperties represent the properties that can be set for scope
units. The fields and their exported types are listed below; for a more
comprehensive explanation of these fields see:
https://www.freedesktop.org/wiki/Software/systemd/dbus/

- `BlockIOAccounting` (bool)
- `BlockIODeviceWeight` ([][]interface{})
- `BlockIOReadBandwidth` ([][]interface{})
- `BlockIOWeight` (uint64)
- `BlockIOWriteBandwidth` ([][]interface{})
- `CPUAccounting` (bool)
- `CPUQuotaPerSecUSec` (uint64)
- `CPUShares` (uint64)
- `CPUUsageNSec` (uint64)
- `ControlGroup` (string)
- `Controller` (string)
- `Delegate` (bool)
- `DeviceAllow` ([][]interface{})
- `DevicePolicy` (string)
- `KillMode` (string)
- `KillSignal` (int32)
- `MemoryAccounting` (bool)
- `MemoryCurrent` (uint64)
- `MemoryLimit` (uint64)
- `Result` (string)
- `SendSIGHUP` (bool)
- `SendSIGKILL` (bool)
- `Slice` (string)
- `StartupBlockIOWeight` (uint64)
- `StartupCPUShares` (uint64)
- `TasksAccounting` (bool)
- `TasksCurrent` (uint64)
- `TasksMax` (uint64)
- `TimeoutStopUSec` (uint64)
