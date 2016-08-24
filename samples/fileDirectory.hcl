param "filename" {
  default = "test"
}

file.directory "render" {
  destination = "{{param `filename`}}"
  force       = true
}
