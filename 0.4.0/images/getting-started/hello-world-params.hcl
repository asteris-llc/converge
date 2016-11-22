param "name" {
  default = "World"
}

file.content "render" {
  destination = "hello.txt"
  content     = "Hello, {{param `name`}}!"
}
