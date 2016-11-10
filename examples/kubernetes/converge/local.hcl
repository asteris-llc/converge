module "cfssl.hcl" "cfssl" {}

module "generate-ca.hcl" "generate-ca" {
  depends = ["module.cfssl"]
}
