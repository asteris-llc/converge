---
title: "unarchive"
slug: "unarchive"
date: "2017-02-22T10:32:05-06:00"
menu:
  main:
    parent: resources
---


Unarchive renders unarchive data


## Example

```hcl
# unarchive

param "zip" {
  default = "/tmp/consul.zip"
}

param "destination" {
  default = "/tmp/consul"
}

task "directory" {
  check = "test -d {{param `destination`}}"
  apply = "mkdir -p {{param `destination`}}"
}

file.fetch "consul.zip" {
  source      = "https://releases.hashicorp.com/consul/0.6.4/consul_0.6.4_linux_amd64.zip"
  destination = "{{param `zip`}}"
}

unarchive "consul.zip" {
  source      = "{{param `zip`}}"
  destination = "{{param `destination`}}"

  depends = ["task.directory", "file.fetch.consul.zip"]
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `source` (required string)

  Source to unarchive - must exist locally

- `destination` (required string)

  Destination for the unarchive - must be a directory

- `hash_type` (optional string)

  HashType of the archive. It is the hash function used to generate the
checksum hash. Valid types are md5, sha1, sha256, and sha512.

- `hash` (optional string)

  Hash of the archive. It is the checksum hash.

- `force` (bool)

  Force indicates whether a file from the unarchived source will replace a
file in the destination if it already exists
If true, the file will be replaced if:
1. no checksum is provided
2. the checksum of the existing file differs from the checksum provided


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `source` (string)

  the source
 
- `destination` (string)

  the destination
 
- `hash_type` (string)

  hash function used to generate the checksum hash of the source; value is
available for lookup if set in the hcl
 
- `hash` (string)

  the checksum hash of the source; value is available for lookup if set in
the hcl
 
- `force` (bool)

  whether a file from the unarchived source will replace a file in the
destination if it already exists
  

