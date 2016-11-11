param "kubernetes-version" {
  default = "1.4.5"
}

module "install-binary.hcl" "kube-apiserver" {
  params {
    url = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-apiserver"
    name = "kube-apiserver"
    destination = "/usr/local/bin/"
  }
}
