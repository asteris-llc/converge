docker.image "busybox" {
  name               = "busybox"
  tag                = "latest"
  inactivity_timeout = "60s"
}
