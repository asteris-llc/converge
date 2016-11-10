module "install-binary.hcl" "cfssl" {
  params {
    url         = "https://pkg.cfssl.org/R1.2/cfssl_linux-amd64"
    name        = "cfssl"
    destination = "/usr/local/bin/"
    working_dir = "/tmp/"
  }
}

module "install-binary.hcl" "cfssljson" {
  params {
    url         = "https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64"
    name        = "cfssljson"
    destination = "/usr/local/bin/"
    working_dir = "/tmp/"
  }
}
