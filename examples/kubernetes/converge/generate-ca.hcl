param "internal-ip" {}

param "ssl-directory" {
  default = "/etc/kubernetes/ssl"
}

param "ca-download-port" {
  default = "9090"
}

param "ca-download-directory" {
  default = "/usr/share/ca"
}

param "ca-archive" {
  default = "/tmp/ca.tar.gz"
}

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

param "nginx-ca-conf-file" {
  default = "/etc/nginx-ca.conf"
}

param "nginx-ca-conf" {
  default = <<EOF
  worker_processes	1;

  events {
    worker_connections	1024;
  }

  http {
    include		mime.types;
    default_type	application/octet-stream;

    sendfile		on;
    keepalive_timeout	65;
    access_log		/dev/stdout;
    error_log     /dev/stderr;

    server {
      listen 80;

      error_page 497 https://$host:$server_port$request_uri;

      location / {
        proxy_connect_timeout	600;
        proxy_send_timeout	600;
        proxy_read_timeout	600;
        send_timeout		600;

        root /usr/share/nginx/html;
        try_files $uri $uri/ =404;
      }
    }
  }
  EOF
}

file.directory "ssl-directory" {
  destination = "{{param `ssl-directory`}}"
  create_all = true
}

file.directory "ca-download-directory" {
  destination = "{{param `ca-download-directory`}}"
  create_all = true
}

file.mode "ca-download-directory" {
  destination = "{{lookup `file.directory.ca-download-directory.destination`}}"
  mode = "0755"
}

file.content "ca-csr" {
  destination = "{{lookup `file.directory.ssl-directory.destination`}}/{{param `ca-csr-file`}}"
  content     = "{{param `ca-csr`}}"
}

file.content "ca-config" {
  destination = "{{lookup `file.directory.ssl-directory.destination`}}/{{param `ca-config-file`}}"
  content     = "{{param `ca-config`}}"
}

task "generate-ca" {
  check = "test -f ca.pem"
  apply = "cat {{lookup `file.content.ca-csr.destination`}}; cfssl gencert -initca {{lookup `file.content.ca-csr.destination`}} | cfssljson -bare ca"
  dir   = "{{lookup `file.directory.ssl-directory.destination`}}"
  depends = ["file.content.ca-csr"]
}

task "archive-ca" {
  check   = "test -f {{lookup `file.directory.ca-download-directory.destination`}}/ca.tar.gz"
  apply   = "tar zcvf {{lookup `file.directory.ca-download-directory.destination`}}/ca.tar.gz *"
  dir     = "{{lookup `file.directory.ssl-directory.destination`}}"
  depends = ["task.generate-ca"]
}

task "copy-ca-archive" {
  check   = "test -f {{param `ca-archive`}}"
  apply   = "cp {{lookup `file.directory.ca-download-directory.destination`}}/ca.tar.gz {{param `ca-archive`}}"
  depends = ["task.archive-ca"]
}

file.content "nginx-ca-conf-file" {
  destination = "{{param `nginx-ca-conf-file`}}"
  content = "{{param `nginx-ca-conf`}}"
}

docker.image "nginx-ca" {
  name = "nginx"
  tag = "1.11.5-alpine"
}

docker.container "nginx-ca" {
  name = "nginx-ca"
  image = "{{lookup `docker.image.nginx-ca.name`}}:{{lookup `docker.image.nginx-ca.tag`}}"
  ports = ["{{param `internal-ip`}}:{{param `ca-download-port`}}:80"]
  volumes = [
    "{{lookup `file.content.nginx-ca-conf-file.destination`}}:/etc/nginx/nginx.conf:ro",
    "{{param `ca-download-directory`}}:/usr/share/nginx/html:ro",
  ]
  depends = ["task.archive-ca"]
}
