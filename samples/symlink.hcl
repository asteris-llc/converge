// this module demonstrates how you use multiple params in a task
param "from" {}

param "to" {}

task "duplicate" {
  check = "readlink {{param `to`}}"
  apply = "ln -s {{param `from`}} {{param `to`}}"
}
