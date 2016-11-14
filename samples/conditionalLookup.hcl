param "str" {
  default = "a"
}

task.query "a" {
  interpreter = "/bin/bash"
  query       = "echo -n {{param `str`}}"
}

switch "lookup-predicate" {
  case "eq `a` `{{lookup `task.query.a.status.stdout`}}`" "lookup" {
    task.query "result" {
      query = "echo 'predicate was true'"
    }
  }

  default {
    task.query "result" {
      query = "echo 'predicate was false'"
    }
  }
}
