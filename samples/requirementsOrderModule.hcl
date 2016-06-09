param "filename" {
  default = "test.txt"
}

param "copyname" {
  default = "test.txt.link"
}

module "basic.hcl" "basic" {
  params {
    filename = "{{param `filename`}}"
  }
}

module "symlink.hcl" "link" {
  params {
    from = "{{param `filename`}}"
    to   = "{{param `copyname`}}"
  }
}
