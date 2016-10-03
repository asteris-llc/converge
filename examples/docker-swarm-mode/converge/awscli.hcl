task "epel-install" {
  check = "test -f /etc/yum.repos.d/epel.repo"
  apply = "yum makecache; yum install -y epel-release"
}

task "pip-install" {
  check   = "yum list installed python-pip"
  apply   = "yum makecache; yum install -y python-pip"
  depends = ["task.epel-install"]
}

task "awscli-install" {
  check   = "which aws"
  apply   = "pip install awscli"
  depends = ["task.pip-install"]
}
