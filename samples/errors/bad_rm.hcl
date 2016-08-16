param "filename" {
  default = "test.txt"
}

task "bad rm" {
  check = "[[ -f {{param `filename`}} ]]"
  apply = "rm -q {{param `filename`}}"
}
