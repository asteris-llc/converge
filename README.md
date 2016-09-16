# Converge

Converge is a configuration management tool.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Converge](#converge)
    - [Installation](#installation)
    - [Usage](#usage)
    - [Development](#development)
        - [Tools](#tools)
        - [RPC](#rpc)
    - [License](#license)

<!-- markdown-toc end -->

## Installation

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

## License

Converge is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for
full details.
