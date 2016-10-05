param "converge-bin-dir" {}

param "image-name" {
  default = "ubuntu"
}

param "image-tag" {
  default = "latest"
}

docker.image "converge" {

}

docker.container "converge" {
  name  = "converge-test"
  image = "{{param `image-name`}}:{{param `image-tag`}}"
  status = "running"

  expose = ["4774"]
  publish_all_ports = "true"

  volumes = [
    "/converge",
    "{{env `PWD`}}/{{param `converge-bin-dir`}}:/converge/bin",
    "{{env `PWD`}}/samples:/converge/samples",
  ]

  command = ["/converge/bin/converge", "server", "--no-token"]
}

task.query "wait" {
  query = "sleep 1"
  depends = ["docker.container.converge"]
}

task.query "rpc-port" {
  query = <<EOF
docker inspect {{lookup `docker.container.converge.name`}} | jq '.[0].NetworkSettings.Ports["4774/tcp"][0].HostPort'
EOF
  depends = ["task.query.wait"]
}

task "tests" {
  check = "docker exec {{lookup `docker.container.converge.name`}} test -f /converge/samples/test.txt"
  apply = "./converge apply /converge/samples/basic.hcl --log-level WARN --rpc-addr :{{lookup `task.query.rpc-port.status.stdout`}}"
  depends = ["task.query.rpc-port"]
}

task "destroy-container" {
  check = "exit $(docker ps -qaf name={{lookup `docker.container.converge.name`}} | wc -l)"
  apply = "docker rm -f {{lookup `docker.container.converge.name`}}"

  depends = ["task.tests"]
}

task "delete-test-txt" {
  check = "test -f ./samples/test.txt"
  apply = "rm -f ./samples/test.txt"

  depends = ["task.destroy-container"]
}
