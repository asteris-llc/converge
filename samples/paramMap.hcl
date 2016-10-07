# ranging a map

param "items" {
  default = {
    a = 1
    b = 2
    c = 3
  }
}

file.content "enumerated" {
  destination = "enumerated.txt"

  content = <<EOF
{{- range $k, $v := paramMap `items` -}}
{{$k}}: {{$v}}
{{end -}}
EOF
}
