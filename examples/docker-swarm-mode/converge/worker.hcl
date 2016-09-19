param "swarm-manager-ip" {}
param "swarm-token-bucket" {}

task.query "swarm-worker-token" {
  query = "aws s3 cp s3://{{param `swarm-token-bucket`}}/worker - | tr -d '\n'"
}

task "swarm-join" {
  check = "docker info 2>/dev/null | grep \"Swarm: active\""
  apply = "docker swarm join --token {{lookup `task.query.swarm-worker-token.status.stdout`}} {{param `swarm-manager-ip`}}:2377"
}
