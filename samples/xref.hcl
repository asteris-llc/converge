param "name" {
  default = "shouldrun.sample"
}

param "dest" {
  default = "{{lookup `file.content.disk.Destination`}}"
}

task "checkspace" {
  interpreter = "/bin/bash"
  check       = "[[ -f {{param `name`}} ]]"
  apply       = "df -h; date > {{param `name`}}"
}

task "finish checkspace" {
  check   = "[[ ! -f {{param `name`}} ]];"
  apply   = "rm -f {{param `name`}}"
  depends = ["task.checkspace"]
}

task "print file info" {
  interpreter = "/bin/bash"
  check       = "[[ ! -f {{lookup `file.content.disk.Destination`}}]]"
  apply       = "rm -f {{param `dest`}}"
}

file.content "disk" {
  destination = "diskspace.txt"
  content     = "{{lookup `task.checkspace.Status.Stdout`}}"
}
