docker.network "test-network" {
  name  = "test-network"
  state = "present"

  labels {
    environment = "test"
  }

  force = true
}
