// this module shows diamond dependencies
task "a" {
  depends = ["task.b", "task.c"]
}

task "b" {
  depends = ["task.d"]
}

task "c" {
  depends = ["task.d"]
}

task "d" {
  depends = []
}
