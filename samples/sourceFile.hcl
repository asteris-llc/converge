param "message" {
  default = "Hello from another module!"
}

module "basic.hcl" "basic" {
  params = {
    message = "{{param `message`}}"
  }
}
