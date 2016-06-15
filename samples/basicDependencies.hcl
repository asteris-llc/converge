param "message" {
  default = "Hello, World!"
}

param "filename" {
  default = "target/test.txt"
}

task "directory" {
  check = "[ -d \"$(dirname {{param `filename`}})\" ] && echo present || (echo absent && exit 1)"
  apply = "mkdir -p $(dirname {{param `filename`}})"
}

task "render" {
  check   = "cat {{param `filename`}} | tee /dev/stderr | grep -q '{{param `message`}}'"
  apply   = "echo '{{param `message`}}' > {{param `filename`}}"
  depends = ["task.directory"]
}
