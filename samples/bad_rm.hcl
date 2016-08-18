param "filename" {
  default = "test.txt"
}

task "bad rm" {
  check = "echo 'check function called'; [[ -f {{param `filename`}} ]]"
  apply = "rm -q {{param `filename`}}"
}
