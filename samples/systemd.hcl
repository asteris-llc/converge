systemd.unit.state "ssh" {
  unit = "ssh.service"
}

file.content "service-info" {
  destination = "out.txt"
  content = <<EOF
Name:        {{lookup `systemd.unit.state.ssh.unit`}}
Path:        {{lookup `systemd.unit.state.ssh.path`}}
Load State:  {{lookup `systemd.unit.state.ssh.loadstate`}}
Type:        {{lookup `systemd.unit.state.ssh.type`}}
Description: {{lookup `systemd.unit.state.ssh.description`}}

{{lookup `systemd.unit.state.ssh.service_properties.BusName`}}
EOF
}
