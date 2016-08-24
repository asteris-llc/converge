---
title: "Resources"
date: "2016-08-24T15:33:42-05:00"

menu:
  main:
    identifier: resources
    weight: 20

---

## Modules

The basic unit of composition in Converge is the module. Modules have parameters
and can call tasks and render templates, among other things. Your usage is
likely to look like this:

```hcl
# install traefik
param "version" {}

module "yum.hcl" "traefik"  {
  params {
    package = "traefik"
    version = "{{param `version`}}"
  }
}
```

You'll invoke this with `converge apply traefik.hcl` to install Traefik from yum
on your system. You can also `converge plan traefik.hcl` to see what changes
will be made before you apply them.

## Using Resources

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
