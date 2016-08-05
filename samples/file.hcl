file.directory "test" {
  destination = "/tmp/test/converge"
  mode        = "777"
  user        = "david"
}

file.absent "absentTest2" {
destination = "/tmp/test/converge2"

}
file.absent "absentTest3" {
destination = "/tmp/test/converge3"

}

file.link "linktest" {
  source      = "/tmp/test/converge"
  destination = "/tmp/test/converge2"
  type = "soft"
}

file.link "hardLinkTest" {
  source      = "/tmp/test/converge"
  destination = "/tmp/test/converge3"
  type = "hard"
}

file.absent "absentTest" {
  destination = "/tmp/hello"
}

file.touch "touchTest1" {
  destination = "/tmp/test/hello1"
}

file.touch "touchTest2" {
  destination = "/tmp/test/hello2"
}

file.touch "touchTest3" {
  destination = "/tmp/test/hello3"
}

file.file "changeMode" {
  directory = "/tmp/test2"
  file = "/tmp/test2/hello"
  mode        = "0776"
  recurse     = true
}
