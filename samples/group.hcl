# create a group, only works on linux
user.group "group" {
  gid   = "123"
  name  = "test"
  state = "present"
}
