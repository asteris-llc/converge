param "ca-csr" {
  default = <<EOF
{
  "CN": "converge-kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "St. Louis",
      "O": "kubernetes",
      "OU": "CA",
      "ST": "Missouri"
    }
  ]
}
EOF
}

param "ca-config" {
  default = <<EOF
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "kubernetes": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "8760h"
      }
    }
  }
}
EOF
}

param "ca-csr-file" {
  default = "ca-csr.json"
}

param "ca-config-file" {
  default = "ca-config.json"
}

file.directory "ssl" {
  destination = "ssl"
}

file.content "ca-csr" {
  destination = "{{lookup `file.directory.ssl.destination`}}/{{param `ca-csr-file`}}"
  content     = "{{param `ca-csr`}}"
}

file.content "ca-config" {
  destination = "{{lookup `file.directory.ssl.destination`}}/{{param `ca-config-file`}}"
  content     = "{{param `ca-config`}}"
}

task "generate-ca" {
  check = "test -f ca.pem"
  apply = "cfssl gencert -initca {{param `ca-csr-file`}} | cfssljson -bare ca"
  dir   = "{{lookup `file.directory.ssl.destination`}}"
}

task "archive-ca" {
  check   = "test -f ca.tar.gz"
  apply   = "tar zcvf ca.tar.gz *"
  dir     = "{{lookup `file.directory.ssl.destination`}}"
  depends = ["task.generate-ca"]
}
