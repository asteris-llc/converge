---
title: Server
date: "2016-08-29T16:49:53-05:00"
menu:
  main:
    parent: converge
    weight: 50

---

Converge comes with a server that can:

- run `plan` and `apply` and stream the results (using
  [gRPC](http://www.grpc.io/))
- serve modules from a given root
- serve the Converge binary itself, for bootstrapping new systems inside your
  network

## HTTPS

You can run the server over HTTPS. If you don't have your own certificates, you
can use [certstrap](https://github.com/square/certstrap) to get some with the
following commands:

```shell
$ certstrap init --common-name your-company
$ certstrap request-cert --ip 127.0.0.1
$ certstrap sign 127.0.0.1 --CA your-company
```

Of course, replace `your-company` and `127.0.0.1` with your company's name and
the your server's IP address, respectively (but those defaults will work fine
for trying it out locally.) The certificates will be placed in `out` in the
directory you run the command from.

Afterwards, reference these files like so:

```shell
converge server --cert-file out/127.0.0.1.crt \
                --key-file out/127.0.0.1.crt \
                --ca-file out/your-company.crt \
                --use-ssl \
                --rpc-token your-token
```

You'll also need to pass the `--ca-file` flag to commands like `plan` and
`apply`, in order to trust your new CA (or put it in the system roots.)

## APIs

Using the Converge command-line interface is good enough for most cases. If you
want to integrate Converge into your system in novel ways, however, an API is
available.

### Authentication

Authentication happens with [JSON Web Tokens](https://jwt.io/). The only
currently supported algorithm is HS512, and issued tokens must have a 30 second
expiration. Tokens are set using the `--rpc-token` [configuration flag]({{< ref
"configuration.md" >}}) to all subcommands that use the API.

### HTTP/2.0 And gRPC

If you want to create your own client for Converge, you'll probably want to use
gRPC. You can get instructions for your chosen langauge in
[the gRPC docs](http://www.grpc.io/docs/), and the protobuf file is
`rpc/pb/root.proto` in the Converge source. If you're using Go, the client
implementations in `rpc/client.go` are your friends.

When using the RPC interface, the JWT token should be sent in the request
metadata's `authorization` field with the prefix `BEARER `.

### HTTP/1.1 And JSON

A pseudo-RESTful interface is available to do the same things the gRPC interface
can do. See the protobuf file for the most up-to-date endpoints and payload
information.

When using the HTTP/1.1 interface, the JWT token should be sent in the
`Authorization` header with the prefix `BEARER`. You can also set the `jwt`
querystring var, or send it in the `jwt` cookie.

## Standalone Server For The Command-Line

The main Converge commands (like `plan` and `apply`) will take a `--local`
argument (or set `CONVERGE_LOCAL=1`.) This will:

1. Start a local RPC server
2. Perform the requested action against the RPC server
3. Shut down the RPC server

During this process, a port (`localhost:26930`) will be opened and RPC will be
running on it. This interface will be protected with an randomly-generated
token, unless you specify `--no-token`

{{< warning title="Don't Disable Tokens" >}}
Please don't disable token generation with `--no-token`. I know we just said you
can, but don't do it. This will open up remote execution of arbitrary
instructions to whoever can reach that port. You can make this process *more*
secure by specifying `--cert-file`, `--key-file`, and optionally `--ca-file` to
connect over HTTPS.
{{< /warning >}}
