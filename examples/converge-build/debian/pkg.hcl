param "name" {}

task "package-install" {
  check = "dpkg -s  {{param `name`}} >/dev/null 2>&1"
  apply = "apt-get update 2>&1 > /dev/null && apt-get -y install {{param `name`}}"
}
