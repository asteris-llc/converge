docker.container "nginx" {
  name  = "nginx"
  image = "nginx:1.10-alpine"
  force = "true"

  network_mode = "bridge"
  networks     = ["test-network", "test"]

  ports = [
    "80",
  ]

  env {
    "FOO" = "BAR"
  }
}
