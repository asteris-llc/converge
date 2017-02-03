## SliceTypeProperties

SliceTypeProperties represent the properties that can be set for slice
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
- `Delegate` (bool)
- `DeviceAllow` ([][]interface{})
- `DevicePolicy` (string)
- `MemoryAccounting` (bool)
- `MemoryCurrent` (uint64)
- `MemoryLimit` (uint64)
- `Slice` (string)
- `StartupBlockIOWeight` (uint64)
- `StartupCPUShares` (uint64)
- `TasksAccounting` (bool)
- `TasksCurrent` (uint64)
- `TasksMax` (uint64)
