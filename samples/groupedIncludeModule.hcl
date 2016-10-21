/* demonstrates using modules containing named groups */
param "jsonfile" {
  default = "test.json"
}

file.content "jsonfile" {
  destination = "{{param `jsonfile`}}"
  content     = "{}"
}

module "groupedModule.hcl" "test1" {
  params = {
    jsonfile = "{{param `jsonfile`}}"
    name     = "test1"
    value    = "test1-val"
  }

  depends = ["file.content.jsonfile"]
}

module "groupedModule.hcl" "test2" {
  params = {
    jsonfile = "{{param `jsonfile`}}"
    name     = "test2"
    value    = "test2-val"
  }

  depends = ["file.content.jsonfile"]
}

module "groupedModule.hcl" "test3" {
  params = {
    jsonfile = "{{param `jsonfile`}}"
    name     = "test3"
    value    = "test3-val"
  }

  depends = ["file.content.jsonfile"]
}
