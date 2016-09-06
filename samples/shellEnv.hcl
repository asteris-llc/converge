param "val" {
  default = "param.foo.value"
}

param "name" {
  default = "FOO"
}

task "refgen" {
  interpreter = "/bin/bash"
  check = "echo -n 'refgen: check'; [[ -f refgen.txt ]]"
  apply = "echo -n 'refgen: apply'; touch refgen.txt"
}

task "shellref" {
  interpreter = "/bin/bash"
  check = "echo 'shellref: {{lookup `task.refgen.Status.Stdout`}}'; [[ -f sr.txt ]]"
  apply = "echo 'stdout: {{lookup `task.refgen.Status.Stdout`}}' | tee sr.txt"
}
