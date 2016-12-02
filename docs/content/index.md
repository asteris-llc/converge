---
title: "Converge"
date: "2016-08-24T15:47:16-05:00"
type: index

menu:
  main:
    identifier: "converge"
    weight: 0
---

Converge is a configuration management tool that makes it easy to manage servers,
laptops and other devices.

Key features:

- Easy to install and run. A single binary and configuration file is all you need.
- A powerful graph engine that automatically generates dependencies and
runs tasks in parallel.
- API-first communication using [grpc](http://grpc.io).
- Module verification: only run trusted modules.

[![Slack Status](http://converge-slack.aster.is/badge.svg)](http://converge-slack.aster.is)
[![Code Climate](https://codeclimate.com/github/asteris-llc/converge/badges/gpa.svg)](https://codeclimate.com/github/asteris-llc/converge)

## Installation

[See the installation documentation.]({{< ref "install.md" >}})

## Usage

See [the docs](http://converge.aster.is)! It's pretty reasonable, though. Here's
a summary:

Converge uses [HCL](https://github.com/hashicorp/hcl) for syntax. HCL is a
superset of JSON that looks (and acts) quite a bit nicer.

The basic unit of composition in converge is the module. Modules have parameters
and contain resources. Creating a module looks something like this:

```hcl
# write "hello world" to disk
param "message" {
  default = "Hello, World in {{param `filename`}}"
}

param "filename" {
  default = "test.txt"
}

file.content "render" {
  destination = "{{param `filename`}}"
  content     = "{{param `message`}}"
}
```

Invoke this with `converge apply --local samples/fileContent.hcl` to install
Traefik from yum on your system. You can also `converge plan --local
samples/fileContent.hcl` to see what changes will be made before you apply them.

## Development

### Tools

For linting, you'll need:

tool | `go get`
---- | --------
 `golint` | github.com/golang/lint/golint
`go tool vet` | (built in)
`gosimple` | honnef.co/go/simple/cmd/gosimple
`unconvert` | github.com/mdempsky/unconvert
`structcheck` | github.com/opennota/check/cmd/structcheck
`varcheck` | github.com/opennota/check/cmd/varcheck
`aligncheck` | github.com/opennota/check/cmd/aligncheck
`gas` | github.com/HewlettPackard/gas

### RPC

You'll need:

- [Google's protobuf compiler](https://github.com/google/protobuf/releases), 3.0
  or above.
- The go protoc plugin: `go get -a github.com/golang/protobuf/protoc-gen-go`
- The grpc gateway plugin(s): `go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger`

### Continuous Integration and Testing

We use Wercker for CI with a custom base image. The Dockerfile for that image
can be found at `/ci/Dockerfile` in the root of the project, and is pushed as
`asteris/converge-ci`. You can test Converge in the container with the
following invocation:

```sh
docker run -i \
           -t \
           --rm \
           --volume $(pwd):/go/src/github.com/asteris-llc/converge \
           asteris/converge-ci \
           /bin/bash -c 'cd /go/src/github.com/asteris-llc/converge; make test'
```

Benchmarks are run with `make bench`, and fuzzing is run with `make
fuzzing/{target}`.

## License

Converge is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for
full details.
