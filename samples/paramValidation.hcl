/*
These are examples of how you can validate your parameters. This includes a simple type checking system, and user-defined constraints.
The constraints are evaulated as text/template fragments, with a small set of funcs included.
*/

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

task.query "sayconverge" {
  interpreter = "/bin/bash"
  query = "echo -n converge"
}

param "converge" {
  default = "{{lookup `task.query.sayconverge.checkstatus.stdout`}}"
  must = ["oneOf `converge`"]
}

param "cipher" {
  default = "Twofish"
  must    = ["oneOf `Rijndael Serpent Twofish`", "notOneOf `DES Blowfish`"]
}
