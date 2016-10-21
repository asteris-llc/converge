param "os" {
  default = "{{platform.OS}}"
}

param "cpu" {
  default = "amd64"
}

param "gopath" {
  default = "{{env `HOME`}}/go"
}

param "go-version" {
  default = "1.7.3"
}

#shasum for amd64
param "go-sha256sum" {
  default = "508028aac0654e993564b6e2014bf2d4a9751e3b286661b0b0040046cf18028e"
}

#arm6l
#param "go-sha256sum" {
#  default = "d02912d121e1455e775a5aa4ecdb2a04f8483ba846e6d2341e1f35b8e507d7b5"
#}


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
    cpu          = "{{param `cpu`}}"
    os           = "{{param `os`}}"

  }

  depends = ["module.deps"]
}

module "protoc.hcl" "protoc" {
  params {
    protoc-version = "{{param `protoc-version`}}"
    cpu            = "x86_64"
    os             = "{{param `os`}}"
  }

  depends = ["module.deps"]
}
