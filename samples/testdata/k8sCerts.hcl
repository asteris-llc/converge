# create certificates for a Kubernetes cluster
# included here for a race condition regression test

param "master_ip" {}

task "directory" {
  check = "test -d ssl"
  apply = "mkdir ssl"
}

task "ca.key" {
  check = "test -f ssl/ca.key"
  apply = "cd ssl; openssl genrsa -out ca.key 2048"

  depends = ["task.directory"]
}

task "ca.crt" {
  check = "test -f ssl/ca.crt"
  apply = "cd ssl; openssl req -x509 -new -nodes -key ca.key -subj \"/CN={{param `master_ip`}}\" -days 10000 -out ca.crt"

  depends = ["task.ca.key"]
}

module "cert.hcl" "server" {
  params {
    cn   = "{{param `master_ip`}}"
    name = "server"
  }

  depends = ["task.ca.key", "task.ca.crt"]
}

module "cert.hcl" "kubelet" {
  params {
    cn   = "{{param `master_ip`}}"
    name = "kubelet"
  }

  depends = ["task.ca.key", "task.ca.crt"]
}
