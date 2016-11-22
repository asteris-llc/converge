param "kubernetes-version" {
  default = "1.4.5"
}

param "kubelet-config-dir" {
  default = "/var/lib/kubelet/"
}

param "kubelet-token" {
  default = "chAng3m3"
}

param "controller-ip" {}

param "api-servers" {}

param "ssl-directory" {
  default = "/etc/kubernetes/ssl"
}

param "cluster-dns" {
  default = "10.32.0.10"
}

param "cluster-domain" {
  default = "cluster.local"
}

package.apt "bridge-utils" {
  name  = "bridge-utils"
  state = "present"
  group = "apt"
}

module "install-binary.hcl" "kubectl" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kubectl"
    name        = "kubectl"
    destination = "/usr/local/bin/"
  }
}

module "install-binary.hcl" "kube-proxy" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kube-proxy"
    name        = "kube-proxy"
    destination = "/usr/local/bin/"
  }
}

module "install-binary.hcl" "kubelet" {
  params {
    url         = "https://storage.googleapis.com/kubernetes-release/release/v{{param `kubernetes-version`}}/bin/linux/amd64/kubelet"
    name        = "kubelet"
    destination = "/usr/local/bin/"
  }
}

file.directory "kubelet-config-dir" {
  destination = "{{param `kubelet-config-dir`}}"
  create_all  = true
}

file.content "kubeconfig" {
  destination = "{{lookup `file.directory.kubelet-config-dir.destination`}}kubeconfig"
  content     = "{{param `kubeconfig`}}"
}

file.content "kubelet-service" {
  destination = "/etc/systemd/system/kubelet.service"
  content     = "{{param `kubelet-service`}}"
  depends     = ["module.kubelet"]
}

task "kubelet-enable" {
  check   = "systemctl is-enabled kubelet"
  apply   = "systemctl daemon-reload; systemctl enable kubelet"
  depends = ["file.content.kubelet-service"]
}

task "kubelet-start" {
  check   = "systemctl is-active kubelet"
  apply   = "systemctl daemon-reload; systemctl start kubelet"
  depends = ["task.kubelet-enable", "package.apt.bridge-utils"]
}

file.content "kube-proxy-service" {
  destination = "/etc/systemd/system/kube-proxy.service"
  content     = "{{param `kube-proxy-service`}}"
  depends     = ["module.kube-proxy"]
}

task "kube-proxy-enable" {
  check   = "systemctl is-enabled kube-proxy"
  apply   = "systemctl daemon-reload; systemctl enable kube-proxy"
  depends = ["file.content.kube-proxy-service"]
}

task "kube-proxy-start" {
  check   = "systemctl is-active kube-proxy"
  apply   = "systemctl daemon-reload; systemctl start kube-proxy"
  depends = ["task.kube-proxy-enable"]
}

param "kubeconfig" {
  default = <<EOF
echo "apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: {{param `ssl-directory`}}/ca.pem
    server: https://{{param `controller-ip`}}:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubelet
  name: kubelet
current-context: kubelet
users:
- name: kubelet
  user:
    token: {{param `kubelet-token`}}
EOF
}

param "kubelet-service" {
  default = <<EOF
[Unit]
Description=Kubernetes Kubelet
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=docker.service
Requires=docker.service

[Service]
ExecStart=/usr/local/bin/kubelet \
  --allow-privileged=true \
  --api-servers={{param `api-servers`}} \
  --cloud-provider= \
  --cluster-dns={{param `cluster-dns`}} \
  --cluster-domain={{param `cluster-domain`}} \
  --configure-cbr0=true \
  --container-runtime=docker \
  --docker=unix:///var/run/docker.sock \
  --network-plugin=cni \
  --network-plugin-dir=/etc/cni/net.d \
  --kubeconfig={{lookup `file.content.kubeconfig.destination`}} \
  --reconcile-cidr=true \
  --serialize-image-pulls=false \
  --tls-cert-file={{param `ssl-directory`}}/kubernetes.pem \
  --tls-private-key-file={{param `ssl-directory`}}/kubernetes-key.pem \
  --v=2

Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}

param "kube-proxy-service" {
  default = <<EOF
[Unit]
Description=Kubernetes Kube Proxy
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
ExecStart=/usr/local/bin/kube-proxy \
  --master=https://{{param `controller-ip`}}:6443 \
  --kubeconfig={{lookup `file.content.kubeconfig.destination`}} \
  --proxy-mode=iptables \
  --v=2

Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
}
