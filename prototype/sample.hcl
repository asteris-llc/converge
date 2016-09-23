switch {
  "1 == 2" {
    task.query "foo" {
      query = "echo foo"
    }
  }
  "2 == 2" {
    task.query "bar" {
      query = "echo bar"
    }
  }
  "3 == 3" {
    task.query "baz" {
      query = "echo baz"
    }
  }
}

task.query "foo" {
  query = "bar"
}

module "foo" "bar.hcl" {
}

switch {
  "1 == 2" {
    task.query "foo" {
      query = "echo foo"
    }
  }
  "2 == 2" {
    task.query "bar" {
      query = "echo bar"
    }
  }
  "3 == 3" {
    task.query "baz" {
      query = "echo baz"
    }
  }
}
