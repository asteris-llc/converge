param "swarm-manager-ip" {}

param "swarm-token-bucket" {}

task "swarm-init" {
  check = "docker info 2>/dev/null | grep \"Swarm: active\""
  apply = "docker swarm init --advertise-addr {{param `swarm-manager-ip`}}"
}

task "swarm-persist-worker-token" {
  check   = "aws s3 ls s3://{{param `swarm-token-bucket`}}/worker"
  apply   = "docker swarm join-token worker -q | aws s3 cp - s3://{{param `swarm-token-bucket`}}/worker"
  depends = ["task.swarm-init"]
}
