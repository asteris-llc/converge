# a generic OpenSSL cert module
# included here for a race condition regression test

param "cn" {
  default = "127.0.0.1"
}

param "name" {
  default = "localhost"
}

param "ca_crt" {
  default = "ca.crt"
}

param "ca_key" {
  default = "ca.key"
}

task "directory" {
  check = "test -d ssl"
  apply = "mkdir ssl"
}

task "key" {
  check   = "test -f ssl/{{param `name`}}.key"
  apply   = "cd ssl; openssl genrsa -out {{param `name`}}.key 2048"
  depends = ["task.directory"]
}

task "csr" {
  check   = "test -f ssl/{{param `name`}}.csr"
  apply   = "cd ssl; openssl req -new -key {{param `name`}}.key --subj \"/CN={{param `cn`}}\" -out {{param `name`}}.csr"
  depends = ["task.key"]
}

task "crt" {
  check   = "test -f ssl/{{param `name`}}.crt"
  apply   = "cd ssl; openssl x509 -req -in {{param `name`}}.csr -CA {{param `ca_crt`}} -CAkey {{param `ca_key`}} -CAcreateserial -out {{param `name`}}.crt -days 10000"
  depends = ["task.csr"]
}
