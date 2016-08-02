param "test1" {
  default = "a,|b,|c,"
}

param "test2" {
  default = "|"
}

file.content "testData" {
  destination = "testData"
  content = "{{range $var, $idx := param `test1` | split (param `test2`) }} {{$idx}}\n{{end}}"
}
