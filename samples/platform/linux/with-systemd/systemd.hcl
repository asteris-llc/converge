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
