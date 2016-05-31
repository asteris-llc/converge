
module "basic.hcl" "basic" {
  message = "Hello from another module!"
}

module "basic.hcl" "advanced" {
  message = "Hello from advanced module!"
  depends = ["basic.hcl.basic"]
}
