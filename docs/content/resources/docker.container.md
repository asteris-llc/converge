---
title: "docker.container"
slug: "docker-container"
date: "2016-12-14T11:24:44-06:00"
menu:
  main:
    parent: resources
---


Container is responsible for creating docker containers. It assumes that
there is already a Docker daemon running on the system.
*Note: docker resources are not currently supported on Solaris.*


## Example

```hcl
/* docker resources are currently not supported on solaris */
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

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


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




## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.
- `name` (string)
  the name of the container
 
- `image` (string)
  the name of the image
 
- `entrypoint` (list of strings)
  the entrypoint into the container
 
- `command` (list of strings)
  the command to run
 
- `workingdir` (string)
  the working directory
 
- `env` (list of strings)
  configured environment variables for the container
 
- `expose` (list of strings)
  additional ports to exposed in the container
 
- `links` (list of strings)
  A list of links for the container in the form of container_name:alias
 
- `portbindings` (list of strings)
  ports to bind
 
- `dns` (list of strings)
  list of DNS servers the container is using
 
- `volumes` (list of strings)
  volumes that have been bind-mounted
 
- `volumesfrom` (list of strings)
  containers from which volumes have been mounted
 
- `publishallports` (bool)
  if true, all ports have been published
 
- `networkmode` (string)
  the mode of the container network
 
- `networks` (list of strings)
  networks the container is connected to
 
- `status` (string)
  the status of the container.
 
- `force` (bool)
  Indicate whether the 'force' flag was set
  

