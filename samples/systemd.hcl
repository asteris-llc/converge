file.content "helloUnitFile" {
  destination = "/tmp/hello.service"

  content = <<END
  [Unit]
  Description=Foo hello world
  [Service]
  ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"

  [Install]
  WantedBy=multi-user.target
END
}

systemd.unit "helloUnit" {
  name    = "/tmp/hello.service"
  active  = "true"
  state   = "linked"
  depends = ["file.content.helloUnitFile"]
}
