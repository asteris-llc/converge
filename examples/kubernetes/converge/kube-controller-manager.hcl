param "kubernetes-version" {
  default = "1.4.5"
}

module "install-binary.hcl" "kube-controller-manager" {
  params {
    url = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-controller-manager"
    name = "kube-controller-manager"
    destination = "/usr/local/bin/"
  }
}
