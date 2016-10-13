param "dir" {
  default = "/tmp/converge"
}

file "dir" {
  destination = "{{param `dir`}}"
  type        = "directory"
  state       = "present"
}

file "hello" {
  destination = "{{param `dir`}}/hello.txt"
  mode        = "0750"
  depends     = ["file.dir"]
}

file "symlink" {
  destination = "{{param `dir`}}/symlink"
  target      = "{{lookup `file.hello.Destination`}}"
  state       = "present"
  type        = "symlink"
}

file "deepdir" {
  destination = "{{param `dir`}}/a/b/c/d/e"
  type        = "directory"
  state       = "present"
}

# create parent dir using force 
file "deepfile" {
  destination = "{{param `dir`}}/a/b/c/d/e/f/g/deep.txt"
  type        = "file"
  state       = "present"
  content     = "converge test"
  force       = "true"
}
