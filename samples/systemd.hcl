param "contents" {
  default = "[Unit]\nDescription=Foo hello world\n[Service]\nExecStart=/bin/bash -c \"while true; do /bin/echo HELLO WORLD; sleep 5; done;\""
}

file.content "render" {
  destination = "/etc/systemd/system/foo.service"
  content     = "{{param `contents`}}"
}

task "reload-daemon" {
  check   = "systemctl is-enabled foo.service | grep -q static"
  apply   = "systemctl daemon-reload"
  depends = ["file.content.render"]
}

systemd.start "startFoo" {
  unit    = "foo.service"
  mode    = "replace"
  depends = ["task.reload-daemon"]
}

systemd.stop "stopFoo" {
  unit    = "foo.service"
  mode    = "replace"
  depends = ["systemd.start.startFoo"]
}
