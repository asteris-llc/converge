param "protoc-version" {
  default = "3.0.2"
}

wait.query "curl-check" {
  check        = "curl --version"
  interval     = "1s"
  max_retry    = 60
  grace_period = "3s"
  interpreter  = "/bin/bash"
}

task "protoc-dl" {
  check       = "[[ -f /usr/local/bin/protoc ]]"
  apply       = "curl -L -o /tmp/protoc-{{param `protoc-version`}}-{{platform.OS}}-x86_64.zip  https://github.com/google/protobuf/releases/download/v{{param `protoc-version`}}/protoc-{{param `protoc-version`}}-{{platform.OS}}-x86_64.zip"
  dir         = "/tmp"
  interpreter = "/bin/bash"
  depends     = ["wait.query.curl-check"]
}

wait.query "unzip-check" {
  check        = "unzip -v"
  interval     = "2s"
  max_retry    = 60
  grace_period = "3s"
  interpreter  = "/bin/bash"
}

task "protoc-extract" {
  check       = "[[ -f /usr/local/bin/protoc ]]"
  apply       = "unzip /tmp/protoc-{{param `protoc-version`}}-linux-x86_64.zip"
  dir         = "/usr/local"
  interpreter = "/bin/bash"
  depends     = ["wait.query.unzip-check", "task.protoc-dl"]
}

file "protoc-link" {
  destination = "/usr/include/google"
  target      = "/usr/local/include/google"
  depends     = ["task.protoc-extract"]
  type        = "symlink"
  interpreter = "/bin/bash"
}
