param "gopath" {
  default = "{{env `HOME`}}/go"
}

param "go-version" {
  default = "1.7.1"
}

param "go-sha256sum" {
  default = "43ad621c9b014cde8db17393dc108378d37bc853aa351a6c74bf6432c1bbd182"
}

param "protoc-version" {
  default = "3.0.0"
}

module "centos/pkg.hcl" "deps" {
  params {
    name = "gcc git which sudo tar unzip"
  }
}

module "go.hcl" "go" {
  params {
    go-version   = "{{param `go-version`}}"
    go-sha256sum = "{{param `go-sha256sum`}}"
  }
  depends = ["module.deps"]
}

module "godeps.hcl" "godeps" {
  params {
    gopath = "{{param `gopath`}}"
  }
   depends = ["module.deps", "module.go"]
}

module "protoc.hcl" "protoc" {
  params {
    protoc-version = "{{param `protoc-version`}}"
  }
  depends = ["module.deps"]
}
