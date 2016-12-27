---
title: "file.fetch"
slug: "file-fetch"
date: "2016-12-22T11:43:14-06:00"
menu:
  main:
    parent: resources
---


Fetch is responsible for fetching files


## Example

```hcl
# fetch files
file.fetch "consul.zip" {
  source      = "https://releases.hashicorp.com/consul/0.6.4/consul_0.6.4_linux_amd64.zip"
  destination = "/tmp/consul.zip"
  hash_type   = "sha256"
  hash        = "abdf0e1856292468e2c9971420d73b805e93888e006c76324ae39416edcf0627"
}

```


## Parameters

Here are the HCL fields that you can specify, along with their expected types
and restrictions:


- `source` (required string)

  Source is the location of the file to fetch

- `destination` (required string)

  Destination for the fetched file

- `hash_type` (optional string)

  HashType is the hash function used to generate the checksum hash

- `hash` (optional string)

  Hash is the checksum hash

- `force` (bool)

  Force indicates whether the file will be fetched if it already exists
If true, the file will be fetched if:
1. no checksum is provided
2. the checksum of the existing file differs from the checksum provided


## Exported Fields

Here are the fields that are exported for use with 'lookup'.  Re-exported fields
will have their own fields exported under the re-exported namespace.


- `source` (string)

  location of the file to fetch
 
- `destination` (string)

  destination for the fetched file
 
- `hash_type` (string)

  hash function used to generate the checksum hash; value is available for
lookup if set in the hcl
 
- `hash` (string)

  the checksum hash; value is available for lookup if set in the hcl
 
- `force` (bool)

  whether the file will be fetched if it already exists
  

