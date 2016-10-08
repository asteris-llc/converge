param "name" {}

task "rpm-install" {
  check = "yum list installed {{param `name`}} >/dev/null 2>&1"
  apply = "yum -y install {{param `name`}}"
}
