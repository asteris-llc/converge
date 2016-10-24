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

  case "true" "unexecuted true" {
    task.query "bar" {
      query = "echo bar"
    }

    file.content "bar-output" {
      destination = "bar-output.txt"
      content     = "{{lookup `task.query.bar.status.stdout`}}"
    }
  }

  case "false" "never exected" {
    task.query "baz" {
      query = "echo baz"
    }

    file.content "baz-output" {
      destination = "baz-output.txt"
      content     = "{{lookup `task.query.baz.status.stdout`}}"
    }
  }
}
