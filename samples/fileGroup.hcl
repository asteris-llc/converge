param "filename" {
  default = "test.txt"
}

file.group "render" {
  destination = "{{param `filename`}}"
  group        = "test"
}
