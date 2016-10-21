/* this module allows you to set a field on the root object of a json file.
   Because this relies on updating a file in place using a temporary file, it
   uses a group to ensure it is not run multiple times in parallel. */

param "jsonfile" {
  default = "test.json"
}

param "name" {
  default = "field"
}

param "value" {
  default = "value"
}

task "update" {
  check = "cat {{param `jsonfile`}} | jq -r -e '.{{param `name`}}'"
  apply = "cat {{param `jsonfile`}} | jq '. + {\"{{param `name`}}\": \"{{param `value`}}\"}' > /tmp/{{param `jsonfile`}} && mv /tmp/{{param `jsonfile`}} {{param `jsonfile`}}"
  group = "groupedModule"
}
