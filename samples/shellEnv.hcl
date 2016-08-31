param "val" {
  default = "param.foo.value"
}

param "name" {
  default = "FOO"
}

task "refgen" {
  check = "echo -n 'refgen: check'; [[ -f refgen.txt ]]"
  apply = "echo -n 'refgen: apply'; touch refgen.txt"
}

task "shellref" {
  #  check = "echo 'stdout: {{lookup `task.refgen.Status.Stdout`}}'; [[ -f sr.txt ]]"
  check = "echo 'check without xref' sr.txt ]]"
  apply = "echo 'stdout: {{lookup `task.refgen.Status.Stdout`}}' | tee sr.txt"
}
