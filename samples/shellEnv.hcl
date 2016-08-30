param "foo" {
  default = "param.foo.value"
}

param "name" {
  default = "FOO"
}

task "Show Env" {
  interpolations = {
    "{{param `name`}}" = "{{param `foo`}}"
  }
  check = "exit 1"
  apply = "echo ${{param `name`}}"
}
