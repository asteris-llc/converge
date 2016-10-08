param "protoc-version" {
  default = "3.0.0"
}

task "install-unzip" {
  check = "yum list installed unzip >/dev/null 2>&1"
  apply = "yum -y install unzip"
}

task "protoc-dl" {
  check = "[[ -f /usr/local/bin/protoc ]]"
  apply = "curl -L -o /tmp/protoc-{{param `protoc-version`}}-{{platform.OS}}-x86_64.zip  https://github.com/google/protobuf/releases/download/v{{param `protoc-version`}}/protoc-{{param `protoc-version`}}-{{platform.OS}}-x86_64.zip"
  dir   = "/tmp"
}

task "protoc-extract" {
  check   = "[[ -f /usr/local/bin/protoc ]]"
  apply   = "sudo unzip /tmp/protoc-{{param `protoc-version`}}-linux-x86_64.zip"
  dir     = "/usr/local"
  depends = ["task.install-unzip", "task.protoc-dl"]
}

file "protoc-link" {
  destination = "/usr/include/google"
  target      = "/usr/local/include/google"
  depends     = ["task.protoc-extract"]
  type        = "symlink"
}
