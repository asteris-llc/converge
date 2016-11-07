---
title: "docker.container"
slug: "docker-container"
date: "2016-11-07T13:35:27-06:00"
menu:
  main:
    parent: resources
---


Container is responsible for creating docker containers. It assumes that
there is already a Docker daemon running on the system.


## Example

```hcl
docker.container "nginx" {
  name  = "nginx"
  image = "nginx:1.10-alpine"
  force = "true"

  ports = [
    "80",
  ]

  env {
    "FOO" = "BAR"
  }
}

```


## Parameters

- `name` (required string)

  name of the container

- `image` (required string)

  the image name or ID to use for the container

- `entrypoint` (list of strings)

  override the container entrypoint

- `command` (list of strings)

  override the container command

- `working_dir` (string)

  override the working directory of the container

- `env` (map of string to string)

  set environment variables in the container

- `expose` (list of strings)

  additional ports to expose in the container

- `links` (list of strings)

  A list of links for the container. Each link entry should be in the form of
container_name:alias

- `ports` (list of strings)

  publish container ports to the host. Each item should be in the following
format:
ip:hostPort:containerPort|ip::containerPort|hostPort:containerPort|containerPort.
Ports can be specified in the format: portnum/proto. If proto is not
specified, "tcp" is assumed

- `dns` (list of strings)

  list of DNS servers for the container to use

- `network_mode` (string)

  the mode of the container network. default: default

- `networks` (list of strings)

  the networks to connect the container to

- `volumes` (list of strings)

  bind mounts volumes

- `volumes_from` (list of strings)

  mounts all volumes from the specified container

- `publish_all_ports` (bool)

  allocates a random host port for all of a container’s exposed ports.
Specified as a boolean value

- `status` (string)


  Valid values: `running` and `created`

  the desired status of the container.

- `force` (bool)

  indicates whether or not the container will be recreated if the state is
not what is expected. By default, the module will only check to see if the
container exists. Specified as a boolean value


