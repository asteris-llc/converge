wait.query "service-health" {
  check        = "nc -z localhost 8080"
  interval     = "1s"
  max_retry    = 10
  grace_period = "1s"
}
