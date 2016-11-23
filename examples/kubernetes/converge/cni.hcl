param "cni_directory" {
  default = "/opt/cni/bin/"
}

param "cni_version" {
  default = "07a8a28637e97b22eb8dfe710eeae1344f69d16e"
}

file.directory "cni_directory" {
  destination = "{{param `cni_directory`}}"
  create_all  = true
}

module "install-binary.hcl" "bridge" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "bridge"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory"]
}

module "install-binary.hcl" "cnitool" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "cnitool"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.bridge"]
}

module "install-binary.hcl" "dhcp" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "dhcp"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.cnitool"]
}

module "install-binary.hcl" "flannel" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "flannel"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.dhcp"]
}

module "install-binary.hcl" "host-local" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "host-local"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.flannel"]
}

module "install-binary.hcl" "ipvlan" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "ipvlan"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.host-local"]
}

module "install-binary.hcl" "loopback" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "loopback"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.ipvlan"]
}

module "install-binary.hcl" "macvlan" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "macvlan"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.loopback"]
}

module "install-binary.hcl" "noop" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "noop"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.macvlan"]
}

module "install-binary.hcl" "ptp" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "ptp"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = false
  }

  depends = ["file.directory.cni_directory", "module.noop"]
}

module "install-binary.hcl" "tuning" {
  params {
    url           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-{{param `cni_version`}}.tar.gz"
    name          = "tuning"
    download_name = "cni.tar.gz"
    extract       = "tar.gz"
    extracted_dir = "bin/"
    destination   = "{{param `cni_directory`}}"
    cleanup       = true
  }

  depends = ["file.directory.cni_directory", "module.ptp"]
}
