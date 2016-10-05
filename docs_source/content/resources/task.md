---
title: "task"
slug: "task"
date: "2016-10-04T13:01:49-05:00"
menu:
  main:
    parent: resources
---


Task allows you to run arbitrary shell commands on your system, first
checking if the command should be run.


## Example

```hcl
param "message" {
  default = "Hello, World!"
}

param "filename" {
  default = "test.txt"
}

task "render" {
  check = "cat {{param `filename`}} | tee /dev/stderr | grep -q '{{param `message`}}'"
  apply = "echo '{{param `message`}}' > {{param `filename`}}"
}

```


## Parameters

- `interpreter` (string)

  the shell interpreter that will be used for your scripts. `/bin/sh` is
used by default.

- `check_flags` (list of strings)

  flags to pass to the `interpreter` binary to check validity. For
`/bin/sh` this is `-n`

- `exec_flags` (list of strings)

  flags to pass to the interpreter at execution time

- `check` (string)

  the script to run to check if a resource needs to be changed. It should
exit with exit code 0 if the resource does not need to be changed, and
1 (or above) otherwise.

- `apply` (string)

  the script to run to apply the resource. Normal shell exit code
expectations apply (that is, exit code 0 for success, 1 or above for
failure.)

- `timeout` (duration string)

  the amount of time the command will wait before halting forcefully. The
format is Go's duraction string. A duration string is a possibly signed
sequence of decimal numbers, each with optional fraction and a unit
suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
"us" (or "µs"), "ms", "s", "m", "h".

- `dir` (string)

  the working directory this command should be run in

- `env` (map of string to string)

  any environment variables that should be passed to the command


