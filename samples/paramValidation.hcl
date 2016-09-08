### Good HCL Examples

param "name" {}

param "password" {
  # type defaults to string
  must = ["len . | lt 5", "len . | gt 25"]
}

param "quorum" {
  type = "int"
  must = ["isOdd", "min 1"]
}

param "blocksize" {
  default = 128

  # type is inferred from default
  must = ["min 50", "max 512"]
}


param "cipher" {
  must = ["oneOf \"Rijndael\" \"Serpent\" \"Twofish\"", "notOneOf \"DES\" \"Blowfish\""]
}
