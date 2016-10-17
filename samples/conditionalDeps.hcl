task.query "foo" {
  query = "echo foo"
}

switch "sample" {
  case "true" "true" {
    file.content "foo-output" {
      destination = "foo-file.txt"
      content     = "{{lookup `task.query.foo.status.stdout`}}"
    }
  }
}
