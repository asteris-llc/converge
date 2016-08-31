param "foo" {
  default = "param.foo.value"
}

param "name" {
  default = "FOO"
}

file.content "xref" {
  destination = "xref"
  contents = "xref"
}

task "Show Env" {
  interpolations = {
    "{{param `name`}}" = "{{param `foo`}}"
  }
  check = "echo '{{lookup `file.content.xref`}}'; exit 1"
  apply = "echo \"{{param `name`}} = ${{param `foo`}}\""
}
