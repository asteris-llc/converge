package.rpm "epel-install" {
  name  = "epel-release"
  state = "present"
}

package.rpm "pip-install" {
  name  = "python-pip"
  state = "present"

  depends = ["package.rpm.epel-install"]
}

task "awscli-install" {
  check = "which aws"
  apply = "pip install awscli"

  depends = ["package.rpm.pip-install"]
}
