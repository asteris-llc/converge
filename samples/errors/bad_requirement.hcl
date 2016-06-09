task "bad_requirement" {
  depends = [ "task.nonexistent" ]
}
