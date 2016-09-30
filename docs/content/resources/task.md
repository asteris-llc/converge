---
title: "task"
slug: "task"
date: "2016-12-16T11:20:35-06:00"
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


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

- `timeout` (optional duration)

  the amount of time the command will wait before halting forcefully.

Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `dir` (string)

  the working directory this command should be run in

- `env` (map of string to string)

  any environment variables that should be passed to the command


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `check` (string)
  the check statement
 
- `apply` (string)
  the apply statement
 
- `dir` (string)
  the working directory of the task
 
- `env` (list of strings)
  environment variables configured for the task
 
- `checkstatus` (CommandResults)
  the status of the check phase
 
- `healthstatus` (resource.HealthStatus)
  the status of the health check
 
- `status` re-exports fields from CommandResults
  the status of the task
  

