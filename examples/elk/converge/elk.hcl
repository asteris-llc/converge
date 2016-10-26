param "user-name" {
  default = "vagrant"
}

module "packages.hcl" "packages" {}

module "docker.hcl" "docker" {
  params = {
    user-name = "{{param `user-name`}}"
  }
}

module "filebeat.hcl" "filebeat" {}

param "elasticsearch-data-directory" {
  default = "/data/elasticsearch"
}

wait.query "elasticsearch-wait" {
  check = <<EOF
status=$(curl -s 'http://localhost:9200/_cluster/health' 2>/dev/null | jq -r .status)
[[ "$status" == "yellow" ]] || [[ "$status" == "green" ]]
EOF

  interval  = "10s"
  max_retry = 10

  depends = ["docker.container.elasticsearch-container"]
}

task "filebeat-elasticsearch-template" {
  check   = "[[  \"$(curl 'http://localhost:9200/_template/filebeat' 2>/dev/null)\" != \"{}\" ]] || exit 1"
  apply   = "curl -XPUT 'http://localhost:9200/_template/filebeat' -d@/etc/filebeat/filebeat.template.json 2>/dev/null"
  depends = ["module.filebeat", "docker.container.elasticsearch-container", "wait.query.elasticsearch-wait"]
}

file.directory "elasticsearch-data-directory" {
  destination = "{{param `elasticsearch-data-directory`}}"
  create_all  = true
}

docker.image "elasticsearch-image" {
  name    = "elasticsearch"
  tag     = "2.4.0"
  depends = ["module.docker"]
}

docker.container "elasticsearch-container" {
  name    = "elasticsearch"
  image   = "{{lookup `docker.image.elasticsearch-image.name`}}:{{lookup `docker.image.elasticsearch-image.tag`}}"
  command = ["elasticsearch", "-Des.insecure.allow.root=true"]
  ports   = ["127.0.0.1:9200:9200"]
  volumes = ["{{param `elasticsearch-data-directory`}}:/usr/share/elasticsearch/data"]
  force   = "true"
  depends = ["file.directory.elasticsearch-data-directory"]
}

docker.image "kibana-image" {
  name    = "kibana"
  tag     = "4.6.0"
  depends = ["module.docker"]
}

docker.container "kibana-container" {
  name  = "kibana"
  image = "{{lookup `docker.image.kibana-image.name`}}:{{lookup `docker.image.kibana-image.tag`}}"
  ports = ["5601:5601"]
  links = ["{{lookup `docker.container.elasticsearch-container.name`}}:elasticsearch"]
  force = "true"
}
