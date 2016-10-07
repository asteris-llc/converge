param "lang" {
  default = ""
}

switch "test-switch" {
  case "eq `spanish` `{{param `lang`}}`" "spanish" {
    file.content "foo-file" {
      destination = "greeting.txt"
      content     = "hola\n"
    }
  }

  case "eq `french` `{{param `lang`}}`" "french" {
    file.content "foo-file" {
      destination = "greeting.txt"
      content     = "salut\n"
    }
  }

  case "eq `japanese` `{{param `lang`}}`" "japanese" {
    file.content "foo-file" {
      destination = "greeting.txt"
      content     = "もしもし\n"
    }
  }

  default {
    file.content "foo-file" {
      destination = "greeting.txt"
      content     = "hello\n"
    }
  }
}
