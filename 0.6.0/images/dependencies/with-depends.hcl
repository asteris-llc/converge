task "names" {
  check = "test -d names"
  apply = "mkdir names"
}

file.content "hello" {
  destination = "names/hello.txt"
  content     = "Hello, World!"

  depends = ["task.names"] # added in the resource that needs the dependency
}
