param "message" {
  default = "Hello, World"
}

param "filename" {
  default = "test.txt"
}

file.content "render" {
  destination = "{{param `filename`}}"
  content     = "{{env `HOME`}}"
}
