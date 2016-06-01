param "message" { default = "Hello, World" }
param "filename" { default = "test.txt" }

template "render" {
  destination = "{{param `filename`}}"
  content = "{{param `message`}}"
}
