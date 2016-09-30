# fetch files
file.fetch "consul.zip" {
  source      = "https://releases.hashicorp.com/consul/0.6.4/consul_0.6.4_linux_amd64.zip"
  destination = "/tmp/consul.zip"
  hash_type   = "sha256"
  hash        = "abdf0e1856292468e2c9971420d73b805e93888e006c76324ae39416edcf0627"
}
