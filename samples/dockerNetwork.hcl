docker.network "test-network" {
  name  = "test-network"
  state = "present"

  labels {
    environment = "test"
  }

  options {
    "com.docker.network.bridge.enable_icc" = "true"
  }

  force = true
}
