/* it would be nice to have a path join utility function. right now you must
pass in params that end with "/" for destination and working directory */

/* install-binary could support wildcards (etcd*) instead of two separate
modules (for etcd and etcdctl, for example). but lack of wildcard support for
file.mode makes this harder */

param "url" {}

param "name" {}

param "download_name" {
  default = "{{param `name`}}"
}

param "destination" {}

param "working_dir" {
  default = "/tmp/"
}

param "extract" {
  default = ""
}

param "extracted_dir" {
  default = "./"
}

param "cleanup" {
  default = true
}

task "download" {
  check = "test -f {{param `destination`}}{{param `name`}} || test -f {{param `download_name`}}"
  apply = "curl --connect-timeout 600 --retry 5 --retry-delay 30 -sSLo {{param `download_name`}} {{param `url`}}"
  dir   = "{{param `working_dir`}}"
}

switch "extract" {
  case "eq `tar.gz` `{{param `extract`}}`" "targz" {
    task "unarchive" {
      check   = "test -f {{param `destination`}}{{param `name`}} || test -d {{param `extracted_dir`}}"
      apply   = "tar xzvf {{param `download_name`}} --no-same-owner -C {{param `working_dir`}}"
      dir     = "{{param `working_dir`}}"
      depends = ["task.download"]
    }
  }

  default {}
}

task "copy-binary" {
  check   = "ls {{param `destination`}}{{param `name`}} > /dev/null 2>&1"
  apply   = "cp {{param `working_dir`}}{{param `extracted_dir`}}{{param `name`}} {{param `destination`}}"
  depends = ["macro.switch.extract", "task.download"]
}

/* wildcard support for file.mode would be nice */
file.mode "set-binary-executable" {
  destination = "{{param `destination`}}{{param `name`}}"
  mode        = "0755"
  depends     = ["task.copy-binary"]
}

task.query "cleanup" {
  query = "echo {{param `cleanup`}}"
}

switch "cleanup" {
  case "{{param `cleanup`}}" "cleanup" {
    task "remove-download" {
      check   = "test ! -f {{param `download_name`}}"
      apply   = "rm {{param `download_name`}}"
      dir     = "{{param `working_dir`}}"
      depends = ["task.copy-binary"]
    }
  }
}

switch "cleanup-extracted" {
  case "and ({{param `cleanup`}}) (ne `./` `{{param `extracted_dir`}}`)" "cleanup-extracted" {
    task "remove-extracted-dir" {
      check   = "test ! -d {{param `extracted_dir`}}"
      apply   = "rm -rf {{param `extracted_dir`}}"
      dir     = "{{param `working_dir`}}"
      depends = ["task.copy-binary"]
    }
  }
}
