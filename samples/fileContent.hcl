param "message" {
  default = "Hello, World in {{param `filename`}}"
}

param "filename" {
  default = "test.txt"
}

file.content "render" {
  destination = "{{param `filename`}}"
  content     = "{{param `message`}}"
}

file.content "render diffs" {
  destination = "{{param `filename`}}.diffs"
  content = "{{lookup `file.content.render.value`}}"
}
