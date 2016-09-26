switch "named-switch" {
  case "a" "eq 1 0" {
    task.query "foo" {
      query = "echo foo"
    }
  }
  case "b" "eq 1 0" {
    task.query "bar" {
      query = "echo bar"
    }
  }
  case "c" "eq 0 1" {
    task.query "baz" {
      query = "echo baz"
    }
  }
}
