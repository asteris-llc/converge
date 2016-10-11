# converge-builder

A converge sample that sets a go build environment for converge. After installing these modules, you should be able
to compile converge.

## Usage

Download converge from the [releases](https://github.com/asteris-llc/converge/releases) page. Then
run the following commands based on your platform:

### Running on CentOS/Redhat/Fedora

```shell
$ converge apply --local converge-centos.hcl
...
```

### Running on Debian/Ubuntu

```shell
$ converge apply --local converge-deb-ubuntu.hcl
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
