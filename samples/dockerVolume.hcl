docker.volume "elasticsearch" {
  name = "elasticsearch"

  labels {
    environment = "test"
  }

  state = "present"
  force = true
}
