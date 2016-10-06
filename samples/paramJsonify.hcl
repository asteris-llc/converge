# this example shows the jsonification of a map, list, and string. jsonify should work on any value.

param "map" {
  default = {
    a = 1
    b = 2
    c = 3
  }
}

param "list" {
  default = [1, 2, 3]
}

param "string" {
  default = "sasquatch"
}

file.content "enumerated" {
  destination = "enumerated.txt"

  content = <<EOF
{
    "map": {{paramMap `map` | jsonify}},
    "list": {{paramList `list` | jsonify}},
    "string": {{param `string` | jsonify}}
}
EOF
}
