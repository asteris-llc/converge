param "kubernetes-version" {
  default = "1.4.5"
}

module "install-binary.hcl" "kubectl" {
  params {
    url = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kubectl"
    name = "kubectl"
    destination = "/usr/local/bin/"
  }
}
