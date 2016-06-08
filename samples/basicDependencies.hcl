param "message" {
  default = "Hello, World!"
}

param "filename" {
  default = "test.txt"
}

task "nothing" {
  check = ""
  apply = ""
}

task "render" {
  check   = "cat {{param `filename`}} | tee /dev/stderr | grep -q '{{param `message`}}'"
  apply   = "echo '{{param `message`}}' > {{param `filename`}} && cat {{param `filename`}}"
  depends = ["nothing"]
}
