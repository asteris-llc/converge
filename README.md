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
your system.

### Writing modules

The content of a simple module is below:

```hcl
# start a systemd unit
param "name" { }

task "start-unit" {
  add = "systemctl start {{param `name`}}"
  remove = "systemctl stop {{param `name`}}"
  result = "echo running"
  status = "systemctl status {{param `name`}}"
}
```

Within a module, you can have tasks. Shown here is a task with the four required
stanzas:

- `add` is run to create the resource controlled by this task
- `remove` is run to remove the same
- `result` returns the ideal state
- `status` returns the actual state

Put simply, if the output of `result` and `status` are not the same (here if the
named unit is not running) `add` will be applied to correct that. It is an error
at the task level if the result is not the same as status after an application.
This task can also be removed, in this instance by stopping the unit.

A module can have multiple tasks. It can also have templates, which render data
to the filesystem from the module's parameters:

```hcl
# create systemd unit file
param "name" { }
param "execStart" { }
param "user" { default = "root" }
param "group" { default = "root" }
param "description" { default = "{{param `name`}}" }
param "targets" { default = [], list = true }

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
  add = "systemctl reload-daemon" # run after content is rendered
  remove = "systemctl reload-daemon" # run after content is removed
  result = ""
  status = "systemd-delta --type=overridden {{param `name`}}"
  depends = ["unit"]
}
```

This module creates a systemd unit file and registers it with the system. Note
that both resource blocks have a name, and that "reload-daemon" depends on
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
