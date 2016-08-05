file "test" {
  destination = "/tmp/test/converge"
  mode = "777"
  user = "david"
  state = "directory"
}

file "linktest" {
  source = "/tmp/test/converge"
  destination = "/tmp/test/converge2"
  state = "link"
}

file "hardLinkTest" {
  source = "/tmp/test/converge"
  destination = "/tmp/test/converge3"
  state = "hard"
}

file "absentTest" {
  destination = "/tmp/hello"
  state = "absent"
}


file "touchTest1" {
  destination = "/tmp/test/hello1"
  state = "touch"
}


file "touchTest2" {
  destination = "/tmp/test/hello2"
  state = "touch"
}
file "touchTest3" {
  destination = "/tmp/test/hello3"
  state = "touch"
}


file "changeMode" {
  destination = "/tmp/test"
  mode = "0776"
  recurse = "true"
  state = "directory"
}
