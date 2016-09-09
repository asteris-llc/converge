### Good HCL Examples

param "name" {
  default = "converge"
}

param "password" {
  # type defaults to string
  default = "password"
  must    = ["len . | lt 5", "len . | gt 25"]
}

param "quorum" {
  default = 3
  must    = ["isOdd", "min 1"]
}

param "blocksize" {
  default = 128

  # type is inferred from default
  must = ["min 50", "max 512"]
}

param "cipher" {
  default = "Twofish"
  must    = ["oneOf \"Rijndael\" \"Serpent\" \"Twofish\"", "notOneOf \"DES\" \"Blowfish\""]
}
