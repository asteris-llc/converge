module "install-binary.hcl" "cfssljson" {
  params {
    url         = "https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64"
    name        = "cfssljson"
    destination = "/usr/local/bin/"
    working_dir = "/tmp/"
  }
}
