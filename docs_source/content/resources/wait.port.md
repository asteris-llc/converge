---
title: "wait.port"
slug: "wait-port"
date: "2016-09-30T10:48:08-04:00"
menu:
  main:
    parent: resources
---




## Example

```hcl
wait.port "8080" {
  host         = "localhost"
  port         = 8080
  protocol     = "tcp"
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

  the amount of time to wait in between checks. The format is Go's duraction
string. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `grace_period` (duration string)

  the amount of time to wait before running the first check and after a
successful check. The format is Go's duraction string. A duration string is
a possibly signed sequence of decimal numbers, each with optional fraction
and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units
are "ns", "us" (or "µs"), "ms", "s", "m", "h".

- `max_retry` (int)

  the maximum number of attempts before the wait fails.


