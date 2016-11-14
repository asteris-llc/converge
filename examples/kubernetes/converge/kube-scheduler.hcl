param "kubernetes-version" {
  default = "1.4.5"
}

module "install-binary.hcl" "kube-scheduler" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-scheduler"
    name        = "kube-scheduler"
    destination = "/usr/local/bin/"
  }
}
