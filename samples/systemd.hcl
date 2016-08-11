systemd.stop "test" {
  unit = "systemd-journald.service"
  mode = "replace"
}

systemd.start "test2" {
  unit = "systemd-journald.service"
  mode = "replace"
}
