/* These tasks demonstrate the use of 'env' and 'dir' for the shell module. It
also demonstrates how the working directory and environment variables can be
populated from params. */

param "working-directory" {
  default = "/tmp/converge-working"
}

param "test-file" {
  default = "test-file"
}

param "var-name" {
  default = "ROLE"
}

param "listen-address" {
  default = ":4001"
}

task "directory" {
  check = "test -d {{param `working-directory`}}"
  apply = "mkdir -p {{param `working-directory`}}"
}

task "touch-file" {
  interpreter = "/bin/bash"
  check_flags = ["-n"]
  dir         = "{{param `working-directory`}}"

  env {
    "{{param `var-name`}}" = "test"
    "ADDRESS"              = "{{param `listen-address`}}"
  }

  check = "cat {{param `test-file`}} | grep \"test: :4001\""
  apply = "echo $ROLE: $ADDRESS > test-file"

  depends = ["task.directory"]
}
