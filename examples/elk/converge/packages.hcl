task "epel-install" {
  check = "test -f /etc/yum.repos.d/epel.repo"
  apply = "yum makecache; yum install -y epel-release"
}

task "jq-install" {
  check   = "yum list installed jq"
  apply   = "yum makecache; yum install -y jq"
  depends = ["task.epel-install"]
}
