module "cfssl.hcl" "cfssl" {}

module "generate-cert.hcl" "generate-cert" {
  depends = ["module.cfssl"]
}
