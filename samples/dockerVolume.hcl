/* docker resources are currently not supported on solaris */
docker.volume "elasticsearch" {
  name = "elasticsearch"

  labels {
    environment = "test"
  }

  state = "present"
  force = true
}
