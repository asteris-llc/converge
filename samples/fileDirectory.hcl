# file.directory will create or ensure a directory is present
file.directory "hello" {
  destination = "hello"
}

# it can also accept a "create_all" parameter to do the equivalent of "mkdir -p"
file.directory "deeper" {
  destination = "deeper/a/b/c"
  create_all  = true
}
