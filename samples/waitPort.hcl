wait.port "8080" {
  host         = "localhost"
  port         = 8080
  interval     = "1s"
  max_retry    = 10
  grace_period = "2s"
}
