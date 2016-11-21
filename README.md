# Converge

Converge is a configuration management tool that makes it easy to manage servers,
laptops and other devices.

Key features:

- Easy to install and run. A single binary and configuration file is all you need.
- A powerful graph engine that automatically generates dependencies and
runs tasks in parallel.
- API-first communication using [grpc](http://grpc.io).
- Module verification: only run trusted modules.

![Converge Graph](examples/docker-swarm-mode/graphs/main.png)

[![Slack Status](http://converge-slack.aster.is/badge.svg)](http://converge-slack.aster.is)
[![Code Climate](https://codeclimate.com/github/asteris-llc/converge/badges/gpa.svg)](https://codeclimate.com/github/asteris-llc/converge)

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Converge](#converge)
    - [Installation](#installation)
    - [Usage](#usage)
    - [Development](#development)
        - [Tools](#tools)
        - [RPC](#rpc)
        - [Continuous Integration and Testing](#continuous-integration-and-testing)
    - [License](#license)

<!-- markdown-toc end -->

## Installation

The `install-converge.sh` script will download and install the converge binary
to your `/usr/local/bin/` directory:

```sh
sudo ./install-converge.sh -v 0.3.0
```

The same installation script is available at `get.converge.sh`:

```shell
curl get.converge.sh | sudo bash -
```

You can also use `go get`:

```sh
go get github.com/asteris-llc/converge
```

or download a release for your platform from the
[releases page on Github](https://github.com/asteris-llc/converge/releases).

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

Invoke this with `converge apply --local samples/fileContent.hcl` to place
a test file on your system. You can also `converge plan --local
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

## License

Converge is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for
full details.
