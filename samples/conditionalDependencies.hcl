# This file contains examples of dependency resolution in

# conditionals. It illustrates the allowed dependencies and is used

# for unit testing.

param "greeting" {
  default = "hello"
}

task.query "say-greeting" {
  interpreter = "/bin/bash"
  query       = "echo -n hello"
}

switch "dependencies" {
  case "eq `hello` `{{lookup `task.query.say-greeting.status.stdout`}}`" "okay" {
    file.content "okay" {
      destination = "/tmp/checked.txt"
      content     = "OK"
    }
  }
}
