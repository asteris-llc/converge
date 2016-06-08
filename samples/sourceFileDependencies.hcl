module "basic.hcl" "basic" {
  params = {
    message = "Hello from another module!"
  }
}

module "basic.hcl" "advanced" {
  params = {
    message = "Hello from advanced module!"
  }

  depends = ["basic"]
}
