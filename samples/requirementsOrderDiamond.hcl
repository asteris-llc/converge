task "a" {
  check = ""
  apply = ""

  depends = ["task.b", "task.c"]
}

task "b" {
  check = ""
  apply = ""

  depends = ["task.d"]
}

task "c" {
  check   = ""
  apply   = ""
  depends = ["task.d"]
}

task "d" {
  check   = ""
  apply   = ""
  depends = []
}
