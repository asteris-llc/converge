# converge-builder

A converge sample that sets a go build environment for converge. After installing these 
modules, you should be able to compile converge.

## Usage

Download converge from the [releases](https://github.com/asteris-llc/converge/releases) page. 
Then run the following commands based on your platform:

### Running on Linux

CentOS, Fedora, Debian and Ubuntu are currently supported.

First install dependencies (go and ProtocolBuffers) as root:

```shell
$ sudo converge apply --local converge-linux.hcl
...
```

Next, install go dependencies as the user who will be building converge

```shell
$ converge apply --local godeps.hcl
...
```

By default, go dependencies will be installed into `$HOME/go`. To override this,
set the `gopath` parameter:

```shell
$ converge apply --local --param gopath=/build/go godeps.hcl
...
```

To run a `make test` of converge, run the `converge-make.hcl` module.
By default the `master` branch will be built, but this can be overridden
using the `branch` parameter.

```shell
$ converge apply --local --param branch=origin/feature/test converge-make.hcl
...
```

## Graphs

### protoc

[protoc.hcl](./protoc.hcl) installs Protocol Buffers.

![prtoc graph](./graphs/protoc.hcl.png)

### go.hcl

[go.hcl](./go.hcl) downloads and installs the go compiler.

![main graph](./graphs/go.hcl.png)

### godeps.hcl

[godeps.hcl](./godeps.hcl) installs the go binaries to `$HOME/go`.

![godeps graph](./graphs/godeps.hcl.png)

### Architecture
The `converge-*` files install packages and import `go.hcl`,
`godeps.hcl` and `protoc.hcl`.

#### Debian/Ubuntu

[converge-deb-ubuntu.hcl](./converge-deb-ubuntu.hcl)

![ubuntu/debian graph](./graphs/converge-deb-ubuntu.hcl.png)

#### CentOS/Redhat/Fedora

[converge-centos.hcl](./converge-centos.hcl)

![centos/redhat/fedora graph](./graphs/converge-centos.hcl.png)
