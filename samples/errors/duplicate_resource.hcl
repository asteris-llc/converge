task "a" {
   check = "[[ -d /tmp/a ]]"
   apply = "mkdir /tmp/a"
}

task "a" {
   check = "[[ -d /tmp/b ]]"
   apply = "mkdir /tmp/b"
}
