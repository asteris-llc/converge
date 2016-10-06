# a (bad) way to range over a list. If you just want each item on a newline, check out paramJoin.hcl.

param "items" {
  default = [1, 2, 3]
}

file.content "enumerated" {
  destination = "enumerated.txt"

  content = <<EOF
{{- range paramList `items` -}}
{{.}}
{{end -}}
EOF
}
