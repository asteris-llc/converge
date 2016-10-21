param "go-version" {
  default = "1.7.3"
}

param "go-sha256sum" {
  default = "508028aac0654e993564b6e2014bf2d4a9751e3b286661b0b0040046cf18028e"
}

param "os" {
  default = "{{platform.OS}}"
}

param "cpu" {
  default = "amd64"
}

file.content "go-sha256" {
  destination = "/tmp/go{{param `go-version`}}-sha256sum.txt"
  content     = "{{param `go-sha256sum`}} go{{param `go-version`}}-sha256sum.txt"
}

task "go-dl" {
  check       = "[[ -f /tmp/go{{param `go-version`}}.{{param `os`}}-{{param `cpu`}}.tar.gz ]]"
  apply       = "curl -L -o /tmp/go{{param `go-version`}}.{{param `os`}}-{{param `cpu`}}.tar.gz  https://storage.googleapis.com/golang/go{{param `go-version`}}.{{param `os`}}-{{param `cpu`}}.tar.gz"
  dir         = "/tmp"
  depends     = ["file.content.go-sha256"]
  interpreter = "/bin/bash"
}

task "go-checksum" {
  check       = "[[ -f /tmp/go{{param `go-version`}}.{{param `os`}}-{{param `cpu`}}.tar.gz ]]"
  apply       = "echo checksum failed"
  dir         = "/tmp"
  depends     = ["file.content.go-sha256", "task.go-dl"]
  interpreter = "/bin/bash"
}

task "go-extract" {
  check       = "[[ -d /usr/local/go ]]"
  apply       = "tar -xvzf /tmp/go{{param `go-version`}}.{{param `os`}}-{{param `cpu`}}.tar.gz 2>&1 >/dev/null"
  dir         = "/usr/local"
  depends     = ["task.go-checksum"]
  interpreter = "/bin/bash"
}

task "go-symlink" {
  check       = "[[ -L /usr/local/bin/go ]]"
  apply       = "ln -s /usr/local/go/bin/go /usr/local/bin/go"
  interpreter = "/bin/bash"
  depends     = ["task.go-extract"]
}
