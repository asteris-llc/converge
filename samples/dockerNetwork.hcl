docker.network "test-network" {
  name  = "test-network"
  state = "present"
  force = true

  labels {
    environment = "test"
  }

  options {
    "com.docker.network.bridge.enable_icc"           = "true"
  }

  ipam_driver = "default"

  ipam_config {
    subnet  = "192.168.129.0/24"
    gateway = "192.168.129.1"

    aux_addresses {
      router  = "192.168.129.40"
      printer = "192.168.129.41"
    }
  }
}
