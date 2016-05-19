# Converge

Converge is a configuration management tool.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Converge](#converge)
    - [Usage](#usage)
        - [Writing modules](#writing-modules)

<!-- markdown-toc end -->

## Usage

Converge uses [HCL](https://github.com/hashicorp/hcl) as a base for
configuration, on top of which it add it's own semantics.

**NOTE: The below documentation is a documentation of the semantics involved
rather than an absolute guide to the system. The commands noted don't work
yet.**

The basic unit of composition in converge is the module. Modules have parameters
and can call tasks and render templates. You're unlikely to have to write a
module yourself, as a consumer of the system. Your usage is more likely to look
this this:

```hcl
# install traefik
param "version" { }

module "yum" "traefik" { }

module "systemd-unit-enabled" "traefik" {
  depends = [ "yum.traefik" ]
}

module "systemd-unit-running" "traefik" {
  depends = [ "yum.traefik" ]
}
```

Invoke this with `converge apply traefik.hcl` to install Traefik from yum on
your system. You can also `converge check traefik.hcl` to check correctness
before you apply.

### Writing modules

The content of a simple module is below:

```hcl
# start a systemd unit
param "name" { }

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

The execution goes something like: if the exit code of `check` is non-zero,
`apply` will be called, then `check`ed again to get the current state.

A module can have multiple tasks. It can also have templates, which render data
to the filesystem from the module's parameters:

```hcl
# create systemd unit file
param "name" { }
param "execStart" { }
param "user" { default = "root" }
param "group" { default = "root" }
param "description" { default = "{{param `name`}}" }
param "targets" { default = [], type = list }

template "unit" {
  destination = "/etc/systemd/system/{{params `name`}}.service"
  content = <<EOF
[Unit]
Description={{param `description`}}
After=network-online.target
Wants=network-online.target
{{range (params `targets`)}}
After={{.}}
Wants={{.}}
{{end}}

[Service]
User={{param `user`}}
Group={{param `group`}}
ExecStart={{param `execStart`}}

[Install]
WantedBy=multi-user.target
EOF
}

task "reload-daemon" {
  check = "systemd-delta --type=overridden {{param `name`}}"
  apply = "systemctl daemon-reload"
  depends = [ "unit" ]
}
```

This module creates a systemd unit file and registers it with the system. Note
that both resource blocks have a name, and that "daemon-reload" depends on
"unit". Converge uses these dependencies to determine build order.

The arguments for the template resource:

- `destination`: where to render the template. If you don't want to render to
  disk, omit this stanza. In that case, you can access the rendered template
  with the `template` template function.
- `content`: this can be a string, a multi-line string (as in the example above)
  or any template function that returns a string. Converge intelligently renders
  the template until there are no more blocks, so sourcing a template from a
  template is fine.

The `task` and `template` blocks are the only ones currently specified. There
may be more low-level resources in the future as we write more modules and find
out where they're needed.
