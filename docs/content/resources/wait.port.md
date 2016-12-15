---
title: "wait.port"
slug: "wait-port"
date: "2016-12-15T15:18:19-06:00"
menu:
  main:
    parent: resources
---




## Example

```hcl
wait.port "8080" {
  host         = "localhost"
  port         = 8080
  interval     = "1s"
  max_retry    = 10
  grace_period = "2s"
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `host` (string)

  a host name or ip address. A TCP connection will be attempted at this host
and the specified Port. default: localhost

- `port` (required int)

  the TCP port to attempt to connect to.

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


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `host` (string)
  the hostname
 
- `port` (int)
  the TCP port
  

