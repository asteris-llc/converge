param "name" {}

task "package-install" {
  check = "rpm -q {{param `name`}} >/dev/null 2>&1"
  apply = "yum -y install {{param `name`}}"
}
