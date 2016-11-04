docker.container "nginx" {
  name  = "nginx"
  image = "nginx:1.10-alpine"
  force = "true"

  network_mode = "bridge"

  ports = [
    "80",
  ]

  env {
    "FOO" = "BAR"
  }
}
