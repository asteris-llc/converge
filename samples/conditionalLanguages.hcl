param "lang" {
  default = ""
}

switch "test-switch" {
  case "eq `spanish` `{{param `lang`}}`" "spanish" {
    file.content "greeting" {
      destination = "greeting.txt"
      content     = "hola"
    }
  }

  case "eq `french` `{{param `lang`}}`" "french" {
    file.content "greeting" {
      destination = "greeting.txt"
      content     = "salut"
    }
  }

  case "eq `japanese` `{{param `lang`}}`" "japanese" {
    file.content "greeting" {
      destination = "greeting.txt"
      content     = "もしもし"
    }
  }

  default {
    file.content "greeting" {
      destination = "greeting.txt"
      content     = "hello"
    }
  }
}
