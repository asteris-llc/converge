---
title: "docker.network"
slug: "docker-network"
date: "2016-11-14T11:12:02-06:00"
menu:
  main:
    parent: resources
---


Network is responsible for managing Docker networks. It assumes that there is
already a Docker daemon running on the system.
*Note: docker resources are not currently supported on Solaris.*


## Example

```hcl
docker.network "test-network" {
  name  = "test-network"
  state = "present"
  force = true

  labels {
    environment = "test"
  }

  options {
    "com.docker.network.bridge.enable_icc" = "true"
  }

  internal    = false
  ipv6        = false
  ipam_driver = "default"

  ipam_config {
    subnet  = "192.168.129.0/24"
    gateway = "192.168.129.1"

    aux_addresses {
      router  = "192.168.129.40"
      printer = "192.168.129.41"
    }
  }
}

```


## Parameters

- `name` (required string)

  name of the network

- `driver` (string)

  network driver. default: bridge

- `labels` (map of string to string)

  labels to set on the network

- `options` (map of string to anything)

  driver specific options

- `ipam_driver` (string)

  ip address management driver. default: default

- `ipam_config` (list of ipamConfigMaps)

  optional custom IPAM configuration. multiple IPAM configurations are
permitted. Each IPAM configuration block should contain one or more of the
following items:

  * subnet:      subnet in CIDR format
  * gateway:     ipv4 or ipv6 gateway for the corresponding subnet
  * ip_range:    container ips are allocated from this sub-ranges (CIDR format)
  * aux_address: auxiliary ipv4 or ipv6 addresses used by the network driver.
                 Aux addresses are specified as a map with a name key and an IP
                 address value

- `internal` (bool)

  restricts external access to the network

- `ipv6` (bool)

  enable ipv6 networking

- `state` (State)


  Valid values: `present` and `absent`

  indicates whether the network should exist. default: present

- `force` (bool)

  indicates whether or not the network will be recreated if the state is not
what is expected. By default, the module will only check to see if the
network exists. Specified as a boolean value


