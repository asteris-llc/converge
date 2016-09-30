param "items" {
  default = [1, 2, 3]
}

file.content "enumerated" {
  destination = "enumerated.txt"
  content     = "{{paramList `items` | join `\n`}}"
}
