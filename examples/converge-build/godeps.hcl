
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
   check        = "[[ -f {{param `go`}} ]]"
   interval     = "1s"
   max_retry    = 60
   grace_period = "3s"
 }

wait.query "git-installed" {
   check        = "which git"
   interval     = "1s"
   max_retry    = 60
   grace_period = "3s"
}

file "gosrc" {
  destination = "{{param `gopath`}}/src/github.com"
  type        = "directory"
  state       = "present"
}

file "gobin" {
  destination = "{{param `gopath`}}/bin"
  type        = "directory"
  state       = "present"
  depends     = ["wait.query.git-installed","wait.query.go-installed"]
}

task "install panicparse" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/pp ]]"
  apply   = "{{param `go`}} get github.com/maruel/panicparse/cmd/pp"
  depends = ["file.gobin","wait.query.git-installed","wait.query.go-installed"]
}

task "install gotool" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check = "[[ -d {{param `gopath`}}/src/github.com/kisielk/gotool ]]"
  apply = "{{param `go` }} get -u github.com/kisielk/gotool"
  depends = ["file.gobin"]
}

task "install golint" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/golint ]]"
  apply   = "{{param `go`}} get -u github.com/golang/lint/golint"
  depends = ["file.gobin", "task.install go guru", "task.install gotool"]
}

task "install goconvey" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/goconvey ]]"
  apply   = "{{param `go`}} get github.com/smartystreets/goconvey"
  depends = ["file.gobin"]
}

task "install goimports" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/goimports ]]"
  apply   = "{{param `go`}} get golang.org/x/tools/cmd/goimports"
  depends = ["file.gobin"]
}

task "install go guru" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/guru ]]"
  apply   = "{{param `go`}} get golang.org/x/tools/cmd/guru"
  depends = ["file.gobin", "task.install goimports"]
}

task "install gosimple" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/gosimple ]]"
  apply   = "{{param `go`}} get honnef.co/go/simple/cmd/gosimple"
  depends = ["file.gobin", "task.install golint"]
}

task "install uconvert" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/unconvert ]]"
  apply   = "{{param `go`}} get github.com/mdempsky/unconvert"
  depends = ["file.gobin", "task.install golint"]
}

task "install structcheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/structcheck ]]"
  apply   = "{{param `go`}} get -u github.com/opennota/check/cmd/structcheck"
  depends = ["file.gobin", "task.install golint"]
}

task "install varcheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/varcheck ]]"
  apply   = "{{param `go`}} get github.com/opennota/check/cmd/varcheck"
  depends = ["file.gobin", "task.install structcheck"]
}

task "install aligncheck" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/aligncheck ]]"
  apply   = "{{param `go`}} get github.com/opennota/check/cmd/aligncheck"
  depends = ["file.gobin", "task.install varcheck"]
}

task "install gas" {
  env {
    "GOPATH" = "{{param `gopath`}}"
  }

  check   = "[[ -f {{param `gopath`}}/bin/gas ]]"
  apply   = "{{param `go`}} get github.com/HewlettPackard/gas"
  depends = ["file.gobin"]
}
