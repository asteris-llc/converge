task.query "foo" {
  query = "echo foo"
}

task.query "bar" {
  query = "echo bar"
}

switch "sample" {
  case "true" "true" {
    task.query "baz" {
      query = "echo baz"
    }

    file.content "baz" {
      destination = "baz-file.txt"
      content     = "{{lookup `task.query.baz.status.stdout`}}"
    }

    file.content "foo-output" {
      destination = "foo-file.txt"
      content     = "{{lookup `task.query.foo.status.stdout`}}"
      depends     = ["task.query.bar", "file.content.baz"]
    }
  }
}
