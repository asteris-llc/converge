param "message" {
  default = "Hello, World"
}

param "filename" {
  default = "test.txt"
}

file.content "dep" {
  destination = "{{param `filename`}}.2"
  content     = "{{param `message`}}"
  depends     = ["file.content.render"]
}

file.content "render" {
  destination = "{{param `filename`}}"
  content     = "{{param `message`}}"
}
