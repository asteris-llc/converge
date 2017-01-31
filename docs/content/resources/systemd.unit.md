---
title: "systemd.unit"
slug: "systemd-unit"
date: "2017-01-31T12:14:13-06:00"
menu:
  main:
    parent: resources
---


UnitState configures loaded systemd units, allowing you to start, stop, or
restart them, reload configuration files, and send unix signals.


## Example

```hcl

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `unit` (required string)

  The name of the unit.  This may optionally include the unit type,
e.g. "foo.service" and "foo" are both valid.

- `state` (string)


    Valid values: `running`, `stopped`, and `restarted`

  The desired state of the unit.  This will affect the current unit job.  Use
`systemd.unit.file` to enable and disable jobs, or `systemd.unit.config` to
set options.

    - `reload` (bool)

  If reload is true then the service will be instructed to reload it's
configuration as if the user had run `systemctl reload`.  This will reload
the actual confguration file for the service, not the systemd unit file
configuration. See `systemctl(1)` for more information.

- `signal_name` (string)


    Only one of `signal_name` or `signal_num` may be set.

  Sends a signal to the process, using it's name.  The signal may be in upper
or lower case (the `SIG` prefix) is optional.  For example, to send user
defined signal 1 to the process you may write any of: usr1, USR1, SIGUSR1,
or sigusr1

see `signal(3)` on BSD/Darwin, or `signal(7)` on GNU Linux systems for more
information on signals

- `signal_number` (uint)


    Only one of `signal_name` or `signal_num` may be set.

  Sends a signal to the process, using it's signal number.  The value must be
an unsigned integer value between 1 and 31 inclusive.


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `unit` (string)


- `state` (string)


- `reload` (bool)


- `signal_name` (string)


- `signal_number` (uint)


- `path` (string)

  The full path to the unit file on disk

- `description` (string)

  Description of the services

- `activestate` (string)

  The active state of the unit

- `loadstate` (string)

  The load state of the unit

- `type` (UnitType)

  The type of the unit

- `status` (string)

  The status represents the current status of the process.  It will be
initialized during planning and updated after apply to reflect any changes.

- `global_properties` (Properties)

  Properties are the global systemd unit properties and will be set for all
unit types.

- `service_properties` (ServiceTypeProperties)

  ServiceProperties contain properties specific to Service unit types

- `SocketProperties` (SocketTypeProperties)

  SocketProperties contain properties specific to Socket unit types

- `DeviceProperties` (DeviceTypeProperties)

  DeviceProperties contain properties specific to Device unit types

- `MountProperties` (MountTypeProperties)

  MountProperties contain properties specific to Mount unit types

- `AutomountProperties` (AutomountTypeProperties)

  AutomountProperties contain properties specific for Autoumount unit types

- `SwapProperties` (SwapTypeProperties)

  SwapProperties contain properties specific to Swap unit types

- `PathProperties` (PathTypeProperties)

  PathProperties contain properties specific to Path unit types

- `TimerProperties` (TimerTypeProperties)

  TimerProperties contain properties specific to Timer unit types

- `SliceProperties` (SliceTypeProperties)

  SliceProperties contain properties specific to Slice unit types

- `ScopeProperties` (ScopeTypeProperties)

  ScopeProperties contain properties specific to Scope unit types
