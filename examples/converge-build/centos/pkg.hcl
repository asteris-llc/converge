param "name" {}

task "rpm-install" {
  check = "rpm -q {{param `name`}} >/dev/null 2>&1"
  apply = "yum -y install {{param `name`}}"
}
