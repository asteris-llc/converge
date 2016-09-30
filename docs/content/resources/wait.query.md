---
title: "wait.query"
slug: "wait-query"
date: "2016-12-16T11:20:35-06:00"
menu:
  main:
    parent: resources
---




## Example

```hcl
wait.query "service-health" {
  check        = "nc -z localhost 8080"
  interval     = "1s"
  max_retry    = 10
  grace_period = "1s"
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `interpreter` (string)

  the shell interpreter that will be used for your scripts. `/bin/sh` is
used by default.

- `check` (required string)

  the script to run to check if a resource is ready. exit with exit code 0 if
the resource is healthy, and 1 (or above) otherwise.

- `check_flags` (list of strings)

  flags to pass to the `interpreter` binary to check validity. For
`/bin/sh` this is `-n`.

- `exec_flags` (list of strings)

  flags to pass to the interpreter at execution time.

- `timeout` (optional duration)

  the amount of time the command will wait before halting forcefully.

Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `dir` (string)

  the working directory this command should be run in.

- `env` (map of string to string)

  any environment variables that should be passed to the command.

- `interval` (optional duration)

  the amount of time to wait in between checks. If the interval is not
specified, it will default to 5 seconds.

Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `grace_period` (optional duration)

  the amount of time to wait before running the first check and after a
successful check. If no grace period is specified, no grace period will be
taken into account.

Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `max_retry` (optional int)

  the maximum number of attempts before the wait fails. If the maximum number
of retries is not set, it will default to 5.


