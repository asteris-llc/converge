param "kubernetes-version" {
  default = "1.4.5"
}

param "kubernetes-config-dir" {
  default = "/var/lib/kubernetes/"
}

param "etcd-servers" {
  default = "https://{{lookup `task.query.internal-ip.status.stdout`}}:2379"
}

param "admin-token" {
  default = "chAng3m3"
}

param "scheduler-token" {
  default = "chAng3m3"
}

param "kubelet-token" {
  default = "chAng3m3"
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

module "install-binary.hcl" "kube-apiserver" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-apiserver"
    name        = "kube-apiserver"
    destination = "/usr/local/bin/"
  }
}

module "install-binary.hcl" "kube-controller-manager" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-controller-manager"
    name        = "kube-controller-manager"
    destination = "/usr/local/bin/"
  }
}

module "install-binary.hcl" "kube-scheduler" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-scheduler"
    name        = "kube-scheduler"
    destination = "/usr/local/bin/"
  }
}

module "install-binary.hcl" "kubectl" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kubectl"
    name        = "kubectl"
    destination = "/usr/local/bin/"
  }
}

file.directory "kubernetes-config-dir" {
  destination = "{{param `kubernetes-config-dir`}}"
  create_all  = true
}

file.content "token-csv" {
  destination = "{{lookup `file.directory.kubernetes-config-dir.destination`}}token.csv"
  content     = "{{param `token-csv`}}"
}

file.content "authorization-policy" {
  destination = "{{lookup `file.directory.kubernetes-config-dir.destination`}}authorization-policy.jsonl"
  content     = "{{param `authorization-policy`}}"
}

file.content "kube-apiserver-service" {
  destination = "/etc/systemd/system/kube-apiserver.service"
  content     = "{{param `kube-apiserver-service`}}"
  depends     = ["module.kube-apiserver"]
}

task "kube-apiserver-enable" {
  check   = "systemctl is-enabled kube-apiserver"
  apply   = "systemctl daemon-reload; systemctl enable kube-apiserver"
  depends = ["file.content.kube-apiserver-service"]
}

task "kube-apiserver-start" {
  check   = "systemctl is-active kube-apiserver"
  apply   = "systemctl daemon-reload; systemctl start kube-apiserver"
  depends = ["task.kube-apiserver-enable"]
}

file.content "kube-controller-manager-service" {
  destination = "/etc/systemd/system/kube-controller-manager.service"
  content     = "{{param `kube-controller-manager-service`}}"
  depends     = ["module.kube-controller-manager"]
}

task "kube-controller-manager-enable" {
  check   = "systemctl is-enabled kube-controller-manager"
  apply   = "systemctl daemon-reload; systemctl enable kube-controller-manager"
  depends = ["file.content.kube-controller-manager-service"]
}

task "kube-controller-manager-start" {
  check   = "systemctl is-active kube-controller-manager"
  apply   = "systemctl daemon-reload; systemctl start kube-controller-manager"
  depends = ["task.kube-controller-manager-enable"]
}

file.content "kube-scheduler-service" {
  destination = "/etc/systemd/system/kube-scheduler.service"
  content     = "{{param `kube-scheduler-service`}}"
  depends     = ["module.kube-scheduler"]
}

task "kube-scheduler-enable" {
  check   = "systemctl is-enabled kube-scheduler"
  apply   = "systemctl daemon-reload; systemctl enable kube-scheduler"
  depends = ["file.content.kube-scheduler-service"]
}

task "kube-scheduler-start" {
  check   = "systemctl is-active kube-scheduler"
  apply   = "systemctl daemon-reload; systemctl start kube-scheduler"
  depends = ["task.kube-scheduler-enable"]
}

param "kube-apiserver-service" {
  default = <<EOF
[Unit]
Description=Kubernetes API Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
ExecStart=/usr/local/bin/kube-apiserver \
  --admission-control=NamespaceLifecycle,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota \
  --advertise-address={{lookup `task.query.internal-ip.status.stdout`}} \
  --allow-privileged=true \
  --apiserver-count=3 \
  --authorization-mode=ABAC \
  --authorization-policy-file={{param `kubernetes-config-dir`}}authorization-policy.jsonl \
  --bind-address=0.0.0.0 \
  --enable-swagger-ui=true \
  --etcd-cafile={{param `ssl-directory`}}/ca.pem \
  --insecure-bind-address=0.0.0.0 \
  --kubelet-certificate-authority={{param `ssl-directory`}}/ca.pem \
  --etcd-servers={{param `etcd-servers`}} \
  --service-account-key-file={{param `ssl-directory`}}/kubernetes-key.pem \
  --service-cluster-ip-range=10.32.0.0/24 \
  --service-node-port-range=30000-32767 \
  --tls-cert-file={{param `ssl-directory`}}/kubernetes.pem \
  --tls-private-key-file={{param `ssl-directory`}}/kubernetes-key.pem \
  --token-auth-file={{lookup `file.content.token-csv.destination`}} \
  --v=2
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}

param "kube-controller-manager-service" {
  default = <<EOF
[Unit]
Description=Kubernetes Controller Manager
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
ExecStart=/usr/local/bin/kube-controller-manager \
--allocate-node-cidrs=true \
--cluster-cidr=10.200.0.0/16 \
--cluster-name=kubernetes \
--leader-elect=true \
--master=http://{{lookup `task.query.internal-ip.status.stdout`}}:8080 \
--root-ca-file={{param `ssl-directory`}}/ca.pem \
--service-account-private-key-file={{param `ssl-directory`}}/kubernetes-key.pem \
--service-cluster-ip-range=10.32.0.0/24 \
--v=2
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}

param "kube-scheduler-service" {
  default = <<EOF
[Unit]
Description=Kubernetes Scheduler
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
ExecStart=/usr/local/bin/kube-scheduler \
  --leader-elect=true \
  --master=http://{{lookup `task.query.internal-ip.status.stdout`}}:8080 \
  --v=2
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}

param "token-csv" {
  default = <<EOF
{{param `admin-token`}},admin,admin
{{param `scheduler-token`}},scheduler,scheduler
{{param `kubelet-token`}},kubelet,kubelet
EOF
}

param "authorization-policy" {
  default = <<EOF
{"apiVersion": "abac.authorization.kubernetes.io/v1beta1", "kind": "Policy", "spec": {"user":"*", "nonResourcePath": "*", "readonly": true}}
{"apiVersion": "abac.authorization.kubernetes.io/v1beta1", "kind": "Policy", "spec": {"user":"admin", "namespace": "*", "resource": "*", "apiGroup": "*"}}
{"apiVersion": "abac.authorization.kubernetes.io/v1beta1", "kind": "Policy", "spec": {"user":"scheduler", "namespace": "*", "resource": "*", "apiGroup": "*"}}
{"apiVersion": "abac.authorization.kubernetes.io/v1beta1", "kind": "Policy", "spec": {"user":"kubelet", "namespace": "*", "resource": "*", "apiGroup": "*"}}
{"apiVersion": "abac.authorization.kubernetes.io/v1beta1", "kind": "Policy", "spec": {"group":"system:serviceaccounts", "namespace": "*", "resource": "*", "apiGroup": "*", "nonResourcePath": "*"}}
EOF
}
