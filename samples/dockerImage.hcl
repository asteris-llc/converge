/* docker resources are currently not supported on solaris */
docker.image "busybox" {
  name               = "busybox"
  tag                = "latest"
  inactivity_timeout = "60s"
}
