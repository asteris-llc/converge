param "internal-ip" {}

param "ssl-directory" {
  default = "/etc/kubernetes/ssl"
}

param "csr-file" {
  default = "csr.json"
}

param "hosts" {
  default = "127.0.0.1,localhost,{{lookup `task.query.hostname.status.stdout`}},{{param `internal-ip`}}"
}

param "ca-url" {
  default = "http://{{param `internal-ip`}}:9090/ca.tar.gz"
}

param "ca-archive" {
  default = "/tmp/ca.tar.gz"
}

param "csr" {
  default = <<EOF
{
  "CN": "kubernetes",
  "hosts": {{param `hosts` | split "," | jsonify}},
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "St. Louis",
      "O": "kubernetes",
      "OU": "cluster",
      "ST": "Missouri"
    }
  ]
}
EOF
}

task.query "hostname" {
  query = "hostname | xargs echo -n"
}

file.directory "ssl" {
  destination = "{{param `ssl-directory`}}"
  create_all  = true
}

file.mode "ssl" {
  destination = "{{lookup `file.directory.ssl.destination`}}"
  mode        = "0700"
}

task "download-ca" {
  check = "test -f {{param `ca-archive`}}"
  apply = "curl --connect-timeout 600 --retry 5 --retry-delay 30 -sSLo {{param `ca-archive`}} {{param `ca-url`}}"
}

task "unarchive-ca" {
  check = "test -f {{lookup `file.directory.ssl.destination`}}/ca.pem"
  apply = "tar xzvf {{param `ca-archive`}} --no-same-owner -C {{lookup `file.directory.ssl.destination`}}"
  depends = ["task.download-ca"]
}

file.content "ca-csr" {
  destination = "{{lookup `file.directory.ssl.destination`}}/{{param `csr-file`}}"
  content     = "{{param `csr`}}"
}

/* I'd like to be able to configure this to run any time file.content.ca-csr
applies (b/c the content has change) rather than checking for the kubernetes.pem
file. right now, I have to manually delete kubernetes.pem if I change the csr
content */
task "generate-cert" {
  check = "test -f kubernetes.pem"

  apply = <<EOF
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -profile=kubernetes \
  {{lookup `file.content.ca-csr.destination`}} | cfssljson -bare kubernetes
EOF

  dir = "{{lookup `file.directory.ssl.destination`}}"

  depends = ["task.unarchive-ca"]
}

task "delete-ca-archive" {
  check   = "test ! -f {{param `ca-archive`}}"
  apply   = "rm {{param `ca-archive`}}"
  depends = ["task.generate-cert"]
}
