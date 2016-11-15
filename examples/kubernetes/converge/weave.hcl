param "cni_directory" {
  default = "/opt/cni/bin/"
}

param "cni_config_directory" {
  default = "/etc/cni/net.d"
}

param "weave_destination" {
  default = "/usr/local/bin/weave"
}

param "peers" {
  default = ""
}

file.directory "cni_directory" {
  destination = "{{param `cni_directory`}}"
  create_all  = true
}

file.directory "cni_config_directory" {
  destination = "{{param `cni_config_directory`}}"
  create_all  = true
}

task "weave_install" {
  check = "test -f /usr/local/bin/weave"
  apply = "curl -L git.io/weave -o {{param `weave_destination`}}"
}

file.mode "weave_binary" {
  destination = "{{param `weave_destination`}}"
  mode        = "0755"
  depends     = ["task.weave_install"]
}

task "weave_setup" {
  check = "test -f /etc/cni/net.d/10-weave.conf"
  apply = "weave setup"

  depends = [
    "file.mode.weave_binary",
    "file.directory.cni_directory",
    "file.directory.cni_config_directory",
  ]
}

task "weave-launch" {
  check   = "weave status"
  apply   = "weave launch {{param `peers`}}"
  depends = ["task.weave_setup"]
}
