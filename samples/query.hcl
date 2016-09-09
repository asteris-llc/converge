query "hostname" {
  query = "hostname"
}

file.content "hostname data" {
  destination = "hostname.txt"
  content     = "{{lookup `query.hostname.status.stdout`}}"
}
