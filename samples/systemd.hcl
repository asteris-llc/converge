systemd.unit "test" {
  name   = "foo.service"
  active = "true"
  state  = "enabled"
}
