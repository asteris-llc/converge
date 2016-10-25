package.rpm "epel-install" {
  name  = "epel-release"
  state = "present"
}

package.rpm "jq-install" {
  name  = "jq"
  state = "present"

  depends = ["package.rpm.epel-install"]
}
