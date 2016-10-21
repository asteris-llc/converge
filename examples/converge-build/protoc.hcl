param "protoc-version" {
  default = "3.0.2"
}

param "os" {
  default = "{{platform.OS}}"
}

param "cpu" {
  default = "x86_64"
}


wait.query "curl-check" {
  check        = "curl -s https://github.com 2>&1 > /dev/null"
  interval     = "2s"
  max_retry    = 60
  grace_period = "3s"
  interpreter  = "/bin/bash"
}

task "protoc-dl" {
  check       = "[[ -f /usr/local/bin/protoc ]]"
  apply       = "curl -L -o /tmp/protoc-{{param `protoc-version`}}-{{param `os`}}-{{param `cpu`}}.zip  https://github.com/google/protobuf/releases/download/v{{param `protoc-version`}}/protoc-{{param `protoc-version`}}-{{param `os`}}-{{param `cpu`}}.zip"
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
  apply       = "unzip /tmp/protoc-{{param `protoc-version`}}-{{param `os`}}-{{param `cpu`}}.zip"
  dir         = "/usr/local"
  interpreter = "/bin/bash"
  depends     = ["wait.query.unzip-check", "task.protoc-dl"]
}

task "protoc-link" {
  check       = "[[ -L /usr/include/google ]]"
  apply       = "ln -s /usr/local/include/google /usr/include/google"
  depends     = ["task.protoc-extract"]
  interpreter = "/bin/bash"
}
