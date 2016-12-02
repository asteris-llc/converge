task.query "hostname" {
  query = "hostname"
}

file.content "hostname-data" {
  destination = "hostname.txt"
  content     = "{{lookup `task.query.hostname.status.stdout`}}"
}
