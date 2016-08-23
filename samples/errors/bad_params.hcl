param "bad" {
  type = "int"
  default = 5
  rule {
    must = [
    	"max 4",
	]
  }
}
