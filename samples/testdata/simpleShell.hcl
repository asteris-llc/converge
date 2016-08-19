# simple shell tasks, for testing out converge tasks
# included here for a race condition regression test

param "working-directory" {
  default = "/tmp/converge-testing"
}

param "test-file" {
  default = "{{param `working-directory`}}/test-file"
}

task "directory" {
  check = "[ -d \"{{param `working-directory`}}\" ] && echo present || (echo absent && exit 1)"
  apply = "mkdir -p {{param `working-directory`}}"
}

task "touch-file" {
  interpreter = "/bin/bash"
  check_flags = ["-n"]

  check   = "cat {{param `test-file`}} | grep \"hello\""
  apply   = "echo hello > {{param `test-file`}}"
  depends = ["task.directory"]
}
