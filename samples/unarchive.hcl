# unarchive

param "zip" {
  default = "/tmp/consul.zip"
}

param "destination" {
  default = "/tmp/consul"
}

task "directory" {
  check = "test -d {{param `destination`}}"
  apply = "mkdir -p {{param `destination`}}"
}

file.fetch "consul.zip" {
  source      = "https://releases.hashicorp.com/consul/0.6.4/consul_0.6.4_linux_amd64.zip"
  destination = "{{param `zip`}}"
}

unarchive "consul.zip" {
  source      = "{{param `zip`}}"
  destination = "{{param `destination`}}"

  depends = ["task.directory", "file.fetch.consul.zip"]
}
