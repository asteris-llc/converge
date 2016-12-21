/* This file demonstrates proper usage of the file.owner module by creating a
 new file, then changing the ownership of that file to a different group. */

file.content "to-change" {
  destination = "tochange"
}

task.query "existing-group" {
  interpreter = "/bin/bash"
  query       = "echo -n $(ls -la {{lookup `file.content.to-change.destination`}} | awk '{print $4}')"
}

task.query "new-group" {
  interpreter = "/bin/bash"
  query       = "echo -n $(groups | xargs -n 1 echo | grep -v $(whoami) | grep -v {{lookup `task.query.existing-group.status.stdout`}} | head -n1)"
}

file.owner "owner-test" {
  destination = "{{lookup `file.content.to-change.destination`}}"
  group       = "{{lookup `task.query.new-group.status.stdout`}}"
}
