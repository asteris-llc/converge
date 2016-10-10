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

module "debian/pkg.hcl" "deps" {
  params {
    name = "apt-utils ca-certificates curl git gcc unzip"
  }
}

module "go.hcl" "go" {
  params {
    go-version   = "{{param `go-version`}}"
    go-sha256sum = "{{param `go-sha256sum`}}"
  }
}

module "godeps.hcl" "godeps" {
  params {
    gopath = "{{param `gopath`}}"
  }
}

module "protoc.hcl" "protoc" {
  params {
    protoc-version = "{{param `protoc-version`}}"
  }
}
