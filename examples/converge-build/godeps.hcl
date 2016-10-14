param "gopath" {
  default = "{{env `HOME`}}/go"
}

param "go" {
  default = "/usr/local/bin/go"
}

param "converge-path" {
  default = "github.com/asteris-llc/converge"
}

wait.query "go-installed" {
  check        = "[[ -L {{param `go`}} ]]"
  interval     = "2s"
  max_retry    = 80
  grace_period = "3s"
  interpreter  = "/bin/bash"
}

wait.query "git-installed" {
  check        = "git version"
  interval     = "3s"
  max_retry    = 60
  grace_period = "3s"
  interpreter  = "/bin/bash"
}

task "gosrc" {
  check       = "[[ -d {{param `gopath`}}/src/github.com ]]"
  apply       = "mkdir -p {{param `gopath`}}/src/github.com"
  interpreter = "/bin/bash"
}

task "gobin" {
  check       = "[[ -d {{param `gopath`}}/bin ]]"
  apply       = "mkdir -p {{param `gopath`}}/bin"
  interpreter = "/bin/bash"
  depends     = ["wait.query.go-installed", "wait.query.git-installed"]
}

task "install panicparse" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/pp ]]"
  apply       = "{{param `go`}} get github.com/maruel/panicparse/cmd/pp"
  depends     = ["task.gobin"]
}

task "install gotool" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -d {{param `gopath`}}/src/github.com/kisielk/gotool ]]"
  apply       = "{{param `go` }} get -u github.com/kisielk/gotool"
  depends     = ["task.gobin"]
}

task "install golint" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/golint ]]"
  apply       = "{{param `go`}} get -u github.com/golang/lint/golint"
  depends     = ["task.gobin", "task.install go guru", "task.install gotool"]
}

task "install goconvey" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/goconvey ]]"
  apply       = "{{param `go`}} get github.com/smartystreets/goconvey"
  depends     = ["task.gobin"]
}

task "install goimports" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/goimports ]]"
  apply       = "{{param `go`}} get golang.org/x/tools/cmd/goimports"
  depends     = ["task.gobin"]
}

task "install go guru" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/guru ]]"
  apply       = "{{param `go`}} get golang.org/x/tools/cmd/guru"
  depends     = ["task.gobin", "task.install goimports"]
}

task "install gosimple" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/gosimple ]]"
  apply       = "{{param `go`}} get honnef.co/go/simple/cmd/gosimple"
  depends     = ["task.gobin", "task.install golint"]
}

task "install uconvert" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/unconvert ]]"
  apply       = "{{param `go`}} get github.com/mdempsky/unconvert"
  depends     = ["task.gobin", "task.install golint"]
}

task "install structcheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/structcheck ]]"
  apply       = "{{param `go`}} get -u github.com/opennota/check/cmd/structcheck"
  depends     = ["task.gobin", "task.install golint"]
}

task "install varcheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/varcheck ]]"
  apply       = "{{param `go`}} get github.com/opennota/check/cmd/varcheck"
  depends     = ["task.gobin", "task.install structcheck"]
}

task "install aligncheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/aligncheck ]]"
  apply       = "{{param `go`}} get github.com/opennota/check/cmd/aligncheck"
  depends     = ["task.gobin", "task.install varcheck"]
}

task "install gas" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  interpreter = "/bin/bash"
  check       = "[[ -f {{param `gopath`}}/bin/gas ]]"
  apply       = "{{param `go`}} get github.com/HewlettPackard/gas"
  depends     = ["task.gobin"]
}
