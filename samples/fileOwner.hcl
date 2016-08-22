param "filename" {
  default = "test.txt"
}

file.owner "render" {
  destination = "{{param `filename`}}"
  user     = "nobody"
}
