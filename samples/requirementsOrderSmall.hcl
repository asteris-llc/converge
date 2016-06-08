task "a" {
  check = ""
  apply = ""

  depends = ["b","c"]
}

task "b" {
  check = ""
  apply = ""

  depends = ["d"]
}

task "c" {
  check = ""
  apply = ""
  depends = ["d"]
}

task "d" {
  check = ""
  apply = ""
  depends = []
}
