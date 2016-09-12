systemd.unit "test" {
  name   = "foo.service"
  active = "true"
  state  = "enabled"

  content = <<EOF
[Unit]
Description=Foo hello world printer
[Service]
ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"

[Install]
WantedBy=multi-user.target
EOF
}
