task "file" {
  check   = "test -f file.txt"
  apply   = "touch file.txt"
  depends = ["module.basic"]
}

module "basic.hcl" "basic" {}
