task "install-tree" {
  check = "dpkg -s tree >/dev/null 2>&1"
  apply = "apt-get install -y tree"
  group = "apt"
}

task "install-jq" {
  check   = "dpkg -s jq >/dev/null 2>&1"
  apply   = "apt-get install -y jq"
  group   = "apt"
  depends = ["task.install-build-essential"]
}

task "install-build-essential" {
  check = "dpkg -s build-essential >/dev/null 2>&1"
  apply = "apt-get install -y build-essential"
  group = "apt"
}
