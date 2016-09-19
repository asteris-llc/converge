param "user-name" {}

module "docker.hcl" "docker" {
  params = {
    user-name = "{{param `user-name`}}"
  }
}

module "awscli.hcl" "awscli" {
  depends = ["module.docker"]
}
