param "filename" {
  default = "test.txt"
}

file.absent "render" {
  destination = "{{param `filename`}}"
}
