task "file1" {
  check = "test -f file1.txt"
  apply = "date > file1.txt"
}

task "lockme1" {
  check = "test -f lockme1.txt"
  apply = "date > lockme1.txt"
  lock  = "mylock"
}

task "lockme2" {
  check = "test -f lockme2.txt"
  apply = "date > lockme2.txt"
  lock  = "mylock"
}

task "lockme3" {
  check = "test -f lockme3.txt"
  apply = "date > lockme3.txt"
  lock  = "mylock"
}

task "file2" {
  check   = "test -f file2.txt"
  apply   = "date > file2.txt"
  depends = ["task.file1"]
}

task "file3" {
  check   = "test -f file3.txt"
  apply   = "date > file3.txt"
  depends = ["lock.unlock.mylock"]
}
