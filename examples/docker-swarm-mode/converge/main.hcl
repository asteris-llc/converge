param "docker-group-user-name" {}

module "docker.hcl" "docker" {
  params = {
    user-name = "{{param `docker-group-user-name`}}"
  }
}

module "awscli.hcl" "awscli" {
  depends = ["module.docker"]
}
