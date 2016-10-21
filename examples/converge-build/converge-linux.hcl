param "gopath" {
  default = "{{env `HOME`}}/go"
}

param "go-version" {
  default = "1.7.3"
}

param "go-sha256sum" {
  default = "508028aac0654e993564b6e2014bf2d4a9751e3b286661b0b0040046cf18028e"
}

param "protoc-version" {
  default = "3.0.2"
}

module "deps.hcl" "deps" {
  params {
    gopath = "{{param `gopath`}}"
    name   = "apt-utils ca-certificates curl make git gcc graphviz unzip"
  }
}

module "go.hcl" "go" {
  params {
    go-version   = "{{param `go-version`}}"
    go-sha256sum = "{{param `go-sha256sum`}}"
  }

  depends = ["module.deps"]
}

module "protoc.hcl" "protoc" {
  params {
    protoc-version = "{{param `protoc-version`}}"
  }

  depends = ["module.deps"]
}
