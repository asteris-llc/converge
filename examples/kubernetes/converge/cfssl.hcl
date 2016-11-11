module "install-binary.hcl" "cfssl" {
  params {
    url         = "https://pkg.cfssl.org/R1.2/cfssl_linux-amd64"
    name        = "cfssl"
    destination = "/usr/local/bin/"
    working_dir = "/tmp/"
  }
}
