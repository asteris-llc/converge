---
title: "systemd.unit.state"
slug: "systemd-unit-state"
date: "2017-03-02T14:16:58-06:00"
menu:
  main:
    parent: resources
---




## Example

```hcl
systemd.unit.state "ssh" {
  unit  = "ssh.service"
  state = "running"
}

systemd.unit.state "acpid" {
  unit = "acpid.socket"
}

file.content "service-info" {
  destination = "out.txt"

  content = <<EOF
Name:        {{lookup `systemd.unit.state.ssh.unit`}}
Path:        {{lookup `systemd.unit.state.ssh.path`}}
Load State:  {{lookup `systemd.unit.state.ssh.loadstate`}}
Type:        {{lookup `systemd.unit.state.ssh.type`}}
Description: {{lookup `systemd.unit.state.ssh.description`}}

Global Properties:
Documentation: {{lookup `systemd.unit.state.ssh.global_properties.Documentation`}}


Service Properties:
AmbientCapabilities: {{lookup `systemd.unit.state.ssh.service_properties.AmbientCapabilities`}}
User:                {{lookup `systemd.unit.state.ssh.service_properties.User`}}
EOF
}

file.content "socket-info" {
  destination = "out.txt"

  content = <<EOF
Name:        {{lookup `systemd.unit.state.acpid.unit`}}
Path:        {{lookup `systemd.unit.state.acpid.path`}}
Load State:  {{lookup `systemd.unit.state.acpid.loadstate`}}
Type:        {{lookup `systemd.unit.state.acpid.type`}}
Description: {{lookup `systemd.unit.state.acpid.description`}}

Global Properties:
Documentation: {{lookup `systemd.unit.state.acpid.global_properties.Documentation`}}

Socket Properties:
Accept:        {{lookup `systemd.unit.state.acpid.socket_properties.Accept`}}
Broadcast:     {{lookup `systemd.unit.state.acpid.socket_properties.Broadcast`}}
Mark:          {{lookup `systemd.unit.state.acpid.socket_properties.Mark`}}
MaxConnection: {{lookup `systemd.unit.state.acpid.socket_properties.MaxConnections`}}
PAMName:       {{lookup `systemd.unit.state.acpid.socket_properties.PAMName`}}
Personality:   {{lookup `systemd.unit.state.acpid.socket_properties.Personality`}}
EOF
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:



## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `unit` (string)

  The name of the unit, including the unit type.
 
- `state` (string)

  The desired state of the unit as configured by the user. It will be one of
`running`, `stopped`, or `restarted` if it was configured by the user, and
an empty string otherwise.
 
- `reload` (bool)

  This field is set to true if the reload flag was configured by the user.
 
- `signal_name` (string)

  The human-readable name of a unix signal that will be sent to the process.
If this is set the name will match the field set in SignalNumber.  See the
man pages for `signal(3)` on BSD/Darwin or `signal(7)` on GNU Linux for a
full explanation of these signals.
 
- `signal_number` (uint)

  The numeric identifier of a unix signal that will be sent to the process.
If this is set it will match the field set in SignalName.  See the man
pages for `signal(3)` on BSD/Darwin or `signal(7)` on GNU Linux for a full
explanation of these signals.
 
- `path` (string)

  The full path to the unit file on disk. This field will be empty if the
unit was not started from a systemd unit file on disk.
 
- `description` (string)

  Description of the services. This field will be empty unless a description
has been provided in the systemd unit file.
 
- `activestate` (string)

  The active state of the unit. It will always be one of: `active`,
`reloading`, `inactive`, `failed`, `activating`, `deactivating`.
 
- `loadstate` (string)

  The load state of the unit.
 
- `type` (UnitType)

  The type of the unit as an enumerated value.  See TypeStr for a human
readable type.
 
- `typestr` (string)

  The type of the unit as a human readable string.  See the man page for
`systemd(1)` for a full description of the types and their meaning.
 
- `status` (string)

  The status represents the current status of the process.  It will be
initialized during planning and updated after apply to reflect any changes.
 
- `global_properties` re-exports fields from Properties

  Properties are the global systemd unit properties and will be set for all
unit types.  See
http://converge.aster.is/extra/systemd-properties/systemd_Properties for
more information.
 
- `service_properties` re-exports fields from ServiceTypeProperties

  ServiceProperties contain properties specific to Service unit types.  This
field is only exported when the unit type is `service`.  See
http://converge.aster.is/extra/systemd-properties/systemd_ServiceTypeProperties
for for more information.
 
- `socket_properties` re-exports fields from SocketTypeProperties

  SocketProperties contain properties specific to Socket unit types. This
field is only exported when the unit type is `socket`. See
http://converge.aster.is/extra/systemd-properties/systemd_SocketTypeProperties
for for more information.
 
- `device_properties` re-exports fields from DeviceTypeProperties

  DeviceProperties contain properties specific to Device unit types. This
field is only exported when the unit type is `device`. See
http://converge.aster.is/extra/systemd-properties/systemd_DeviceTypeProperties
for for more information.
 
- `mount_properties` re-exports fields from MountTypeProperties

  MountProperties contain properties specific to Mount unit types. This field
is only exported when the unit type is `mount`. See
http://converge.aster.is/extra/systemd-properties/systemd_MountTypeProperties
for for more information.
 
- `automount_properties` re-exports fields from AutomountTypeProperties

  AutomountProperties contain properties specific for Autoumount unit
types. This field is only exported when the unit type is
`automount`. http://converge.aster.is/extra/systemd-properties/systemd_AutomountTypeProperties
for for more information.
 
- `swap_properties` re-exports fields from SwapTypeProperties

  SwapProperties contain properties specific to Swap unit types. This field
is only exported when the unit type is
`swap`. http://converge.aster.is/extra/systemd-properties/systemd_SwapTypeProperties
for for more information.
 
- `path_properties` re-exports fields from PathTypeProperties

  PathProperties contain properties specific to Path unit types. This field
is only exported when the unit type is
`path`. http://converge.aster.is/extra/systemd-properties/systemd_PathTypeProperties
for for more information.
 
- `timer_properties` re-exports fields from TimerTypeProperties

  TimerProperties contain properties specific to Timer unit types. This field
is only exported when the unit type is
`timer`. http://converge.aster.is/extra/systemd-properties/systemd_TimerTypeProperties
for for more information.
 
- `slice_properties` re-exports fields from SliceTypeProperties

  SliceProperties contain properties specific to Slice unit types. This field
is only exported when the unit type is
`slice`. http://converge.aster.is/extra/systemd-properties/systemd_SliceTypeProperties
for for more information.
 
- `scope_properties` re-exports fields from ScopeTypeProperties

  ScopeProperties contain properties specific to Scope unit types. This field
is only exported when the unit type is
`scope`. http://converge.aster.is/extra/systemd-properties/systemd_ScopeTypeProperties
for for more information.
  

