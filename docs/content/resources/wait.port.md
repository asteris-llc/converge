---
title: "wait.port"
slug: "wait-port"
date: "2016-10-03T10:23:34-04:00"
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

- `host` (string)

  a host name or ip address. A TCP connection will be attempted at this host
and the specified Port.

- `port` (required int)

  the TCP port to attempt to connect to.

- `interval` (duration string)

  the amount of time to wait in between checks. The format is Go's duration
string. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". If
the interval is not specified, it will default to 5 seconds.

- `grace_period` (duration string)

  the amount of time to wait before running the first check and after a
successful check. The format is Go's duration string. A duration string is
a possibly signed sequence of decimal numbers, each with optional fraction
and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units
are "ns", "us" (or "µs"), "ms", "s", "m", "h". If no grace period is
specified, no grace period will be taken into account.

- `max_retry` (int)

  the maximum number of attempts before the wait fails. If the maximum number
of retries is not set, it will default to 5.


