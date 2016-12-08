---
title: "docker.image"
slug: "docker-image"
date: "2016-12-08T15:04:22-06:00"
menu:
  main:
    parent: resources
---


Image is responsible for pulling Docker images. It assumes that there is
already a Docker daemon running on the system.
*Note: docker resources are not currently supported on Solaris.*


## Example

```hcl
/* docker resources are currently not supported on solaris */
docker.image "busybox" {
  name               = "busybox"
  tag                = "latest"
  inactivity_timeout = "60s"
}

```


## Parameters

- `name` (required string)

  name of the image to pull


- `tag` (string)

  tag of the image to pull. default: latest


- `inactivity_timeout` (duration)

  the amount of time to wait after a period of inactivity. The timeout is
reset each time new data arrives.

Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".



## Exported Fields
- `name` (string)
  name of the image
 
- `tag` (string)
  tag of the image
  

