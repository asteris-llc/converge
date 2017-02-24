switch "linux-only" {
  case "{{eq `linux` (platform.OS)}}" "linux" {
    task.query "ok" {
      query = "echo 'OK!'"
    }
  }

  default {
    error "not-linux" {
      error = "this module is only supported on Linux systems"
    }
  }
}
