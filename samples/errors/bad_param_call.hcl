# bad param call in the task below. This should produce an error.
task "bad" {
  check = "{{param `nonexistent`}}"
}
