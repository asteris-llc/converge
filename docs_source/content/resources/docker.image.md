---
title: "docker.image"
slug: "docker-image"
date: "2016-09-08T23:18:03-07:00"
menu:
  main:
    parent: resources
---


Image is responsible for pulling Docker images. It assumes that there is
already a Docker daemon running on the system.


## Example

```hcl
docker.image "busybox" {
  name               = "busybox"
  tag                = "latest"
  inactivity_timeout = "60s"
}

```


## Parameters

- `name` (string)

  name of the image to pull

- `tag` (string)

  tag of the image to pull

- `inactivity_timeout` (duration_string)

  the amount of time to wait after a period of inactivity. The timeout is
reset each time new data arrives. The format is Go's duration string. A
duration string is a possibly signed sequence of decimal numbers, each with
optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".


