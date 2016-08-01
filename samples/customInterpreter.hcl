task "testruby" {
  interpreter = "/usr/bin/ruby"
  syntaxcheck = "-c"
  check = "puts 'check func'; exit(1.kind_of? Integer)"
  apply = "puts 'apply func'"
}