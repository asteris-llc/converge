param "name" {}

switch "deps" {
  case "or (eq `{{platform.LinuxDistribution}}` `centos`) (eq `{{platform.LinuxDistribution}}` `fedora`)" "redhat-family" {
    package.rpm "deps" {
      name  = "coreutils gcc git graphviz make perl-Digest-SHA sudo tar unzip which"
      state = "present"
    }
  }

  case "or (eq `{{platform.LinuxDistribution}}` `debian`) (eq `{{platform.LinuxDistribution}}` `ubuntu`) (eq `{{platform.LinuxDistribution}}` `raspbian`)" "debian-family" {
    task "package-install" {
      check = "dpkg -s {{param `name`}} >/dev/null 2>&1"
      apply = "apt-get update 2>&1 > /dev/null && apt-get -y install {{param `name`}}"
    }
  }
}
