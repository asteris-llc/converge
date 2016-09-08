---
title: "task"
slug: "task"
date: "2016-08-31T09:51:20-05:00"
menu:
  main:
    parent: resources
---


Task.query allows you to run arbitrary shell commands on your system and store
the output without taking any actions that would have permanent effects.  It
overlaps significantly with the Task module, differing in the use of a field,
`query` in place of `check` and `apply`, and in the fact that there is never an
application phase.

## Example

```hcl
task.query "hostname" {
  query = "hostname"
}

file.content "hostname data" {
  destination = "hostname.txt"
  content     = "{{lookup `task.query.hostname.Status.Stdout`}}"
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

- `query` (string)

  the script to run to query the system.  The output and exit status is saved to
  be accessed by other modules.

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


## Exported Fields

The following fields are available from other modules via `lookup`

- `CheckStmt` (string)

  the script that will be executed during planning.  It is the script provided
  to `check` with any template references resolved.

- `ApplyStmt` (string)

  the script that will be executed during application.  It is the script
  provided to `apply` with any template references resolved.

- `Dir` (string)

  the working directory of the script.  It will be an empty string if unset.

- `Env` (list of strings)

  a list of environment variables passed into the script with `env` in the form
  of `key=value`. If no environment variables were set then it returns an empty
  list.

- `Status` (command results)

  the execution status of the last task that was run.  During planning this will
  be the result of `check`, and during application it will be the result of
  `apply`.

- `CheckStatus` (command results)

  the status of the initial `check` run.  During planning this is equivivlent to
  `Status`, during application this will provide access to the results of the
  initial call to `check`.

### Command Results

The command results structure provides fields related to the execution status of
a task.  `Status` and `CheckStatus` both return command results type fields that
allow you to access any of these defined fields.

- `ExitStatus` (unsigned integer)

  the exit code of the process. The meaning of the fields is system dependent.

- `Stdout` (string)

  contains all data written to stdout by the process.

- `Stderr` (string)

  contains all data written to stderr by the process.
