# This sample demonstrates how to do value passing with tasks

param "val" {
  default = "param.foo.value"
}

param "name" {
  default = "FOO"
}

task "refgen" {
  interpreter = "/bin/bash"
  check       = "echo -n '{{param `name`}}'; [[ -f refgen.txt ]]"
  apply       = "echo -n '{{param `name`}}.apply'; touch refgen.txt"
}

task "shellref" {
  check = "echo 'shellref.check: {{lookup `task.refgen.Status.ExitStatus`}}'; [[ -f refgen.txt.2 ]]"
  apply = "echo 'shellref.apply: {{lookup `task.refgen.Status.Stdout`}}'; touch refgen.txt.2"
}

task "cleanup" {
  interpreter = "/bin/bash"
  check       = " [[ ! ( -f refgen.txt || -f refgen.txt.2 ) ]] "
  apply       = "rm -f refgen.txt refgen.txt.2"
  depends     = ["task.refgen", "task.shellref"]
}
