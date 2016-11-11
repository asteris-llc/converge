param "etcd-url" {
  default = "https://github.com/coreos/etcd/releases/download/v3.0.14/etcd-v3.0.14-linux-amd64.tar.gz"
}

param "etcd-destination" {
  default = "/usr/local/bin/"
}

module "install-binary.hcl" "etcd" {
  params {
    url           = "{{param `etcd-url`}}"
    name          = "etcd"
    download_name = "etcd.tar.gz"
    destination   = "{{param `etcd-destination`}}"
    working_dir   = "/tmp/"
    extract       = "tar.gz"
    extracted_dir = "etcd-v3.0.14-linux-amd64/"
    cleanup       = "false"
  }
}
