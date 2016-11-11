param "etcd-initial-cluster" {
  # default = "{{lookup `task.query.hostname.status.stdout`}}=https://{{lookup `task.query.internal-ip.status.stdout`}}:2380 "
}

param "ssl-directory" {
  default = "/etc/kubernetes/ssl"
}

task.query "hostname" {
  query = "hostname | xargs echo -n"
}

task.query "internal-ip" {
  query = "ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/' | xargs echo -n"
}

file.directory "etcd-config-dir" {
  destination = "{{param `ssl-directory`}}"
  create_all  = true
}

file.directory "etcd-data-dir" {
  destination = "/var/lib/etcd"
  create_all  = true
}

file.content "etcd-service" {
  destination = "/etc/systemd/system/etcd.service"
  content = "{{param `etcd-service`}}"
}

task "etcd-enable" {
  check   = "systemctl is-enabled etcd"
  apply   = "systemctl daemon-reload; systemctl enable etcd"
  depends = ["file.content.etcd-service"]
}

task "etcd-start" {
  check   = "systemctl is-active etcd"
  apply   = "systemctl daemon-reload; systemctl start etcd"
  depends = ["task.etcd-enable"]
}

param "etcd-service" {
  default = <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/coreos

[Service]
ExecStart=/usr/local/bin/etcd --name {{lookup `task.query.hostname.status.stdout`}} \
  --cert-file={{param `ssl-directory`}}/kubernetes.pem \
  --key-file={{param `ssl-directory`}}/kubernetes-key.pem \
  --peer-cert-file={{param `ssl-directory`}}/kubernetes.pem \
  --peer-key-file={{param `ssl-directory`}}/kubernetes-key.pem \
  --trusted-ca-file={{param `ssl-directory`}}/ca.pem \
  --peer-trusted-ca-file={{param `ssl-directory`}}/ca.pem \
  --initial-advertise-peer-urls https://{{lookup `task.query.internal-ip.status.stdout`}}:2380 \
  --listen-peer-urls https://{{lookup `task.query.internal-ip.status.stdout`}}:2380 \
  --listen-client-urls https://{{lookup `task.query.internal-ip.status.stdout`}}:2379,http://127.0.0.1:2379 \
  --advertise-client-urls https://{{lookup `task.query.internal-ip.status.stdout`}}:2379 \
  --initial-cluster-token etcd-cluster-0 \
  --initial-cluster {{param `etcd-initial-cluster`}} \
  --initial-cluster-state new \
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}
