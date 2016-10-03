param "swarm-manager-ip" {}

param "swarm-token-bucket" {}

wait.port "swarm-manager-port" {
  host         = "{{param `swarm-manager-ip`}}"
  port         = 2377
  interval     = "10s"
  max_retry    = 60
  grace_period = "2s"
}

wait.query "wait-for-swarm-worker-token" {
  check        = "aws s3 ls s3://{{param `swarm-token-bucket`}}/worker"
  interval     = "10s"
  max_retry    = 30
  grace_period = "5s"
  depends      = ["wait.port.swarm-manager-port"]
}

task.query "swarm-worker-token" {
  query   = "aws s3 cp s3://{{param `swarm-token-bucket`}}/worker - | tr -d '\n'"
  depends = ["wait.query.wait-for-swarm-worker-token"]
}

task "swarm-join" {
  check = "docker info 2>/dev/null | grep \"Swarm: active\""
  apply = "docker swarm join --token {{lookup `task.query.swarm-worker-token.status.stdout`}} {{param `swarm-manager-ip`}}:2377"
}
