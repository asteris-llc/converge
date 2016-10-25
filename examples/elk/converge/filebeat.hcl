param "filebeat-service" {
  default = "filebeat"
}

task.query "beats-repo-key" {
  query = "rpm --import https://packages.elastic.co/GPG-KEY-elasticsearch"
}

file.content "beats-repo" {
  destination = "/etc/yum.repos.d/beats.repo"

  content = <<EOF
[beats]
name=Elastic Beats Repository
baseurl=https://packages.elastic.co/beats/yum/el/$basearch
enabled=1
gpgkey=https://packages.elastic.co/GPG-KEY-elasticsearch
gpgcheck=1
EOF

  depends = ["task.query.beats-repo-key"]
}

package.rpm "filebeat-install" {
  name  = "filebeat"
  state = "present"

  depends = ["file.content.beats-repo"]
}

file.content "filebeat-yml" {
  destination = "/etc/filebeat/filebeat.yml"

  content = <<EOF
filebeat:
  prospectors:
    - paths:
        - /var/log/*.log
        - /var/log/messages
      input_type: log
  registry_file: /var/lib/filebeat/registry
output:
  elasticsearch:
    hosts: ["localhost:9200"]
EOF

  depends = ["package.rpm.filebeat-install"]
}

task "filebeat-enable" {
  check   = "systemctl is-enabled {{param `filebeat-service`}}"
  apply   = "systemctl enable {{param `filebeat-service`}}"
  depends = ["file.content.filebeat-yml"]
}

task "filebeat-start" {
  check   = "systemctl is-active {{param `filebeat-service`}}"
  apply   = "systemctl start {{param `filebeat-service`}}"
  depends = ["task.filebeat-enable"]
}
