docker.image "nginx" {
  name    = "nginx"
  tag     = "latest"
  timeout = "60s"
}

docker.container "nginx" {
  name    = "nginx"
  image   = "nginx:latest"
  depends = ["docker.image.nginx"]
}
