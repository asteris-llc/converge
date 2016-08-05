# Converge

Converge is a configuration management tool.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Converge](#converge)
    - [Installation](#installation)
    - [Usage](#usage)
        - [Writing Modules](#writing-modules)
        - [Built-in Modules](#built-in-modules)
            - [File Modules](#file-modules)
                - [File (file)](#file)
                - [Owner (fille.owner)](#owner-fileowner)
                - [Mode (file.mode)](#mode-filemode)
        - [Server](#server)
            - [Module Hosting](#module-hosting)
            - [Binary Hosting](#binary-hosting)
    - [License](#license)

<!-- markdown-toc end -->

## Installation

```sh
go get github.com/asteris-llc/converge
```

or download a release for your platform from the
[releases page on Github](https://github.com/asteris-llc/converge/releases).

## Usage

Converge uses [HCL](https://github.com/hashicorp/hcl) as a base for
configuration, on top of which it add it's own semantics.

The basic unit of composition in converge is the module. Modules have parameters
and can call tasks and render templates, among other things. Your usage is
likely to look this this:

```hcl
# install traefik
param "version" {}

module "yum" "traefik" {}

module "systemd-unit-enabled" "traefik" {
  requires = ["yum.traefik"]
}

module "systemd-unit-running" "traefik" {
  requires = ["yum.traefik"]
}
```

Invoke this with `converge apply traefik.hcl` to install Traefik from yum on
your system. You can also `converge plan traefik.hcl` to see what changes will
be made before you apply them.

### Writing Modules

The content of a simple module is below:

```hcl
# start a systemd unit
param "name" {}

task "start-unit" {
  check = "systemctl status {{param `name`}} | tee /dev/stderr | grep -q running"
  apply = "systemctl start {{param `name`}}"
}
```

Within a module, you can have tasks. Shown here is a task with two stanzas:

- `check` returns the actual state, and an error code indicating if it needs to
  be changed
- `apply` is run to create the resource controlled by this task. You can omit
  this stanza if you want the command to be purely informational.

If the exit code of `check` is non-zero, `apply` will be called, after which
time `check` will be called again to get the new state and success.

A module can have multiple tasks. It can also have templates, which render data
to the filesystem from the module's parameters:

```hcl
# create systemd unit file
param "name" {}

param "execStart" {}

param "user" {
  default = "root"
}

param "group" {
  default = "root"
}

param "description" {
  default = "{{param `name`}}"
}

file.content "unit" {
  destination = "/etc/systemd/system/{{param `name`}}.service"

  content = <<EOF
[Unit]
Description={{param `description`}}
After=network-online.target
Wants=network-online.target

[Service]
User={{param `user`}}
Group={{param `group`}}
ExecStart={{param `execStart`}}

[Install]
WantedBy=multi-user.target
EOF
}

task "reload-daemon" {
  check    = "systemd-delta --type=overridden {{param `name`}}"
  apply    = "systemctl daemon-reload"
  requires = ["file.content.unit"]
}
```

This module creates a systemd unit file and registers it with the system. Note
that both resource blocks have a name, and that "daemon-reload" depends on
"unit". Converge uses these dependencies to determine execution order.

The arguments for the `file.content` resource:

- `destination`: where to render the template. If you don't want to render to
  disk, omit this stanza. In that case, you can access the rendered template
  with the `template` template function.
- `content`: this can be a string, a multi-line string (as in the example above)
  or any template function that returns a string. Converge intelligently renders
  the template until there are no more blocks, so sourcing a template from a
  template is fine.

### Built-in Modules

Converge ships with a number of built-in modules. These can be used for common
tasks without having to write your own `task` declarations.

#### File Modules

The `file` module performs basic operations on files as well as setting attributes
like mode and owner.
It takes a variety of parameters
| parameter | choices   | comments                                             |
| ----------| :-------: | --------                                             |
| state     | file, link, directory, hard, touch, absent| If `directory`, all immediate subdirectores will be created if they do not exist. If `file`, the file will NOT be created if it does not exist, instead the mode and owner of the file will be checked. If `link`, a symbolic link will be created. Use `hard` for hard links. If `absent`, directories will be recursively deleted. If `touch` an empty file will be created if not a file is not alread present.  |
| destination |         | path to file being managed.                          |
| source    |           | if state is `link` or `hard`, this is the source of the link |
| recurse   |           | if state is `directory`, recursively set the file attributes|
| mode      |           | the mode this file should have.                                |
| user      |           | the username of the user this file should belong to.  |

Sample:
```hcl
file "test" {
  destination = "/tmp/test/converge"
  mode = "777"
  user = "david"
  state = "directory"
}

file "linktest" {
  source = "/tmp/test/converge"
  destination = "/tmp/test/converge2"
  state = "link"
}

file "hardLinkTest" {
  source = "/tmp/test/converge"
  destination = "/tmp/test/converge3"
  state = "hard"
}

file "absentTest" {
  destination = "/tmp/hello"
  state = "absent"
}


file "touchTest1" {
  destination = "/tmp/test/hello1"
  state = "touch"
}


file "touchTest2" {
  destination = "/tmp/test/hello2"
  state = "touch"
}
file "touchTest3" {
  destination = "/tmp/test/hello3"
  state = "touch"
}


file "changeMode" {
  destination = "/tmp/test"
  mode = "0776"
  recurse = "true"
  state = "directory"
}

```


##### File (file)

The `file.owner` module takes two required parameters:

- `destination`: the file whose permissions should be checked
- `user`: the username of the user this file should belong to

Sample:

```hcl
file.owner "test" {
  destination = "test.txt"
  user        = "david"
}
```

##### Owner (file.owner)

The `file.owner` module takes two required parameters:

- `destination`: the file whose permissions should be checked
- `user`: the username of the user this file should belong to

Sample:

```hcl
file.owner "test" {
  destination = "test.txt"
  user        = "david"
}
```


##### Mode (file.mode)

The `file.mode` module takes two required parameters:

- `destination`: the file whose permissions should be checked
- `mode`: the octal mode of the file

Sample:

```hcl
file.mode "test" {
  destination = "test.txt"
  mode        = "0644"
}
```

### Server

Converge can run a server to serve modules for bootstrapping. It can also host
itself for downloading onto new systems. Use `converge server` to start it. See
the `--help` for how to enable HTTPS.

#### Module Hosting

Modules are hosted at `/modules/`, which is a simple directory listing (control
with `--root`, set to the current working directory by default), and will be
made available publically. Use these modules by running `converge plan
http://your.host:8080/modules/yourModule.hcl` or similar.

#### Binary Hosting

Converge can also host its own binary. This is turned off by default, and is
enabled with the `--self-serve` flag to `converge server`. You can use this to
bootstrap a new system without downloading the relevant version of converge over
an external connection. It will be available at
`http://your.host:8080/bootstrap/binary`.

## License

Converge is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for
full details.
