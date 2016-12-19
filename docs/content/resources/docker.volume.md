---
title: "docker.volume"
slug: "docker-volume"
date: "2016-12-19T14:58:57-06:00"
menu:
  main:
    parent: resources
---


Volume is responsible for managing Docker volumes. It assumes that there is
already a Docker daemon running on the system.
*Note: docker resources are not currently supported on Solaris.*


## Example

```hcl
/* docker resources are currently not supported on solaris */
docker.volume "elasticsearch" {
  name = "elasticsearch"

  labels {
    environment = "test"
  }

  state = "present"
  force = true
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `name` (required string)

  name of the volume

- `driver` (string)

  volume driver. default: local

- `labels` (map of string to string)

  labels to set on the volume

- `options` (map of string to string)

  driver specific options

- `state` (State)


	Valid values: `present` and `absent`

  indicates whether the volume should exist.

- `force` (bool)

  indicates whether or not the volume will be recreated if the state is not
what is expected. By default, the module will only check to see if the
volume exists. Specified as a boolean value


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `name` (string)

  volume name
 
- `labels` (map of string to string)

  volume labels
 
- `driver` (string)

  driver the volume is configured to use
 
- `options` (map of string to string)

  driver-specific options
 
- `state` (State)

  volume state
 
- `force` (bool)

  reflects whether or not the force option was configured
  

