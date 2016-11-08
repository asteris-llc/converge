docker.container "nginx" {
  name  = "nginx"
  image = "nginx:1.10-alpine"
  force = "true"

  ports = [
    "80",
  ]

  env {
    "FOO" = "BAR"
  }
}
