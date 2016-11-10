/* it would be nice to have a path join utility function. right now you must
pass in params that end with "/" for destination and working directory */

param "url" {}

param "name" {}

param "destination" {}

param "working_dir" {}

param "extract" {
  default = ""
}

task "download-binary" {
  check = "test -f {{param `destination`}}{{param `name`}}"
  apply = "curl -SLo {{param `working_dir`}}{{param `name`}} {{param `url`}}"
}

task "move-binary" {
  check   = "test -f {{param `destination`}}{{param `name`}}"
  apply   = "mv {{param `working_dir`}}{{param `name`}} {{param `destination`}}{{param `name`}}"
  depends = ["task.download-binary"]
}

file.mode "set-binary-executable" {
  destination = "{{param `destination`}}{{param `name`}}"
  mode        = "0755"
  depends     = ["task.move-binary"]
}
