param "image" {
  default = "nginx"
}

param "image-tag" {
  default = "1.10-alpine"
}

param "container" {
  default = "nginx-server"
}

docker.image "nginx" {
  name    = "{{param `image`}}"
  tag     = "{{param `image-tag`}}"
  timeout = "60s"
}

docker.container "nginx" {
  name  = "{{param `container`}}"
  image = "{{param `image`}}:{{param `image-tag`}}"
  force = true

  expose = [
    "80",
    "443/tcp",
    "8080",
  ]

  publish_all_ports = false

  ports = [
    "80",
  ]

  env {
    "FOO" = "BAR"
  }

  dns = ["8.8.8.8", "8.8.4.4"]

  depends = ["docker.image.nginx"]
}
