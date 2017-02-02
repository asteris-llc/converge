# unarchive

param "destination" {
  default = "/tmp/consul"
}

task "directory" {
  check = "test -d {{param `destination`}}"
  apply = "mkdir -p {{param `destination`}}"
}

unarchive "consul.zip" {
  source      = "https://releases.hashicorp.com/consul/0.6.4/consul_0.6.4_linux_amd64.zip"
  destination = "{{param `destination`}}"

  depends = ["task.directory"]
}
