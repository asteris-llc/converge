param "converge-bin-dir" {}

module "docker-test.hcl" "ubuntu-xenial" {
  params {
    image-name = "ubuntu"
    image-tag = "xenial"
    converge-bin-dir = "{{param `converge-bin-dir`}}"
  }
}

module "docker-test.hcl" "centos-7" {
  params {
    image-name = "centos"
    image-tag = "7"
    converge-bin-dir = "{{param `converge-bin-dir`}}"
  }
}

module "docker-test.hcl" "debian-jessie" {
  params {
    image-name = "debian"
    image-tag = "jessie"
    converge-bin-dir = "{{param `converge-bin-dir`}}"
  }
}
