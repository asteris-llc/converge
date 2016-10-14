param "val" {
  default = "2"
}

switch "test-switch" {
  case "eq 1 {{param `val`}}" "a" {
    file.content "foo-file" {
      destination = "foo-file.txt"
      content     = "{{lookup `file.content.foo1.content`}}"
    }

    file.content "foo-file-2" {
      destination = "foo-file-2.txt"
      content     = "{{lookup `file.content.foo-file.destination`}}"
    }
  }

  case "eq 3 1" "b" {
    task.query "baz" {
      query = "echo baz"
    }
  }
}

file.content "foo1" {
  destination = "foo1.txt"
  content     = "foo1\n"
}

file.content "foo3" {
  destination = "foo2.txt"
  content     = "{{lookup `file.content.foo1.destination`}}\n"
}
