---
title: "Using Dependencies"
date: "2016-09-09T00:13:59-05:00"

menu:
  main:
    parent: "converge"
    weight: 20
---

In the [getting started guide]({{< ref "getting-started.md" >}}) we talked about
dependencies, and briefly mentioned that they're *super important* for Converge
to work properly. But we didn't really go into them there&hellip; so here we
are!

## Graph Walking

Briefly, Converge operates by thinking about your deployment as a graph of tasks
that need to be done. It walks around the graph from the leaves (AKA tasks with
no dependencies) all the way to to the root (AKA Rome, where all roads lead.)
Let's explore what that means with a graph from the [params example in the
getting started]({{< ref "getting-started.md" >}}#params) guide:

{{< figure src="/images/getting-started/hello-world-params.png"
           caption="A graph with a parameter. The file hello.txt depends on the name parameter." >}}

What does Converge do when you ask it to apply this graph? Converge will load
the file and start walking at a node that has no dependencies. In this case,
that's `param.name = "World"`. When `param.name` has been successfully walked,
Converge will move on to `File: hello.txt`, and then to the root (`/`). Once
each node is marked as successful, our graph walk is complete. Neat!

## The Graph Command

All the graphs we've seen so far have just been the output of Converge's `graph`
command. When asked, Converge will load up any modules you specify and render
them as [Graphviz](http://graphviz.org/) dot output. You can render like so:

```bash
$ converge graph --local yourModule.hcl | dot -Tpng > yourModule.png
```

When you're developing modules, make a habit of rendering them as graphs. It
makes it easier to think about how the graph will be executed.

## Cross-Node References

Resources may reference one another as long as the references do not introduce
circular dependencies. When creating a reference from one node to another, we
can use the `lookup` command to reference fields of another entry's module. The
available fields will vary depending on the module and should be documented
along with each module. The example below illustrates using `lookup` to access
fields from a `docker.image` node from within `docker.container`:

```hcl
docker.image "nginx" {
  name    = "nginx"
  tag     = "1.10-alpine"
  timeout = "60s"
}

docker.container "nginx" {
  name  = "nginx-server"
  image = "{{lookup `docker.image.nginx.name`}}:{{lookup `docker.image.nginx.tag`}}"
  force = "true"
  expose = [
    "80",
    "443/tcp",
    "8080",
  ]
  publish_all_ports = "false"
  ports = [
    "80",
  ]
  env {
    "FOO" = "BAR"
  }
  dns = ["8.8.8.8", "8.8.4.4"]
}
```

As we can see, `lookup` syntax resembles that of parameters and adds implicit
dependencies between nodes.

## Explicit Dependencies

When we're walking our graph, there are a lot of operations that can be done in
parallel. For this to work, you will need to specify dependencies between
resources in the same file. Let's take the following example:

```hcl
task "names" {
  check = "test -d names"
  apply = "mkdir names"
}

file.content "hello" {
  destination = "names/hello.txt"
  content     = "Hello, World!"
}
```

As a human reading this, we can clearly see that `file.content.hello` is
dependent on `task.names`, because the file needs the directory to be created
before it can write files into it. But Converge doesn't know that yet, so here's
how it looks:

{{< figure src="/images/dependencies/without-depends.png"
           caption="The graph output of the above module. Converge hasn't connected the directory and file." >}}

To fix this, we'll need to specify `depends` on our `file.content`. `depends` is
a list of resources in the current module that must be successfully walked
before walking the containing resource. They're specified as the resource type,
a dot, then the resource name. So `task "names"` above becomes `task.names`.

```hcl
task "names" {
  check = "test -d names"
  apply = "mkdir names"
}

file.content "hello" {
  destination = "names/hello.txt"
  content     = "Hello, World!"

  depends = ["task.names"] # added in the resource that needs the dependency
}
```

Now Converge correctly sees that it needs to walk `task.names` before
`file.content.hello`:

{{< figure src="/images/dependencies/with-depends.png"
           caption="The graph output of the above module. Converge now sees the dependency between the directory and the file." >}}

{{< note title="Future Improvements" >}}
We're working hard on making Converge better at detecting situations like this
automatically. Ideally, you wouldn't have to specify dependencies at all, and it
would all work like [the param example in the getting started guide]({{< ref
"getting-started.md" >}}#params). We're not quite there yet, but keep an eye
out!
{{< /note >}}

## Grouping

There can be some scenarios where a group of tasks are not explicitly dependent
on each other but also cannot be run in parallel. A good example of this is
package management tools like
[apk](http://wiki.alpinelinux.org/wiki/Alpine_Linux_package_management) or
[apt](https://wiki.debian.org/Apt). As an example, let's look at this file which
installs three packages:

```hcl
task "install-tree" {
  check = "dpkg -s tree >/dev/null 2>&1"
  apply = "apt-get install -y tree"
}

task "install-jq" {
  check = "dpkg -s jq >/dev/null 2>&1"
  apply = "apt-get install -y jq"
}

task "install-build-essential" {
  check = "dpkg -s build-essential >/dev/null 2>&1"
  apply = "apt-get install -y build-essential"
}
```

Here is what the corresponding graph looks like:

{{< figure src="/images/dependencies/without-groups.png" caption="The graph output of the above module. Converge will attempt to run each task in parallel." >}}

If you were to execute apply against this graph, you would end up with errors
that look something like this:

```bash
E: Could not get lock /var/lib/apt/lists/lock - open (11: Resource temporarily unavailable)
E: Unable to lock directory /var/lib/apt/lists/
```

This is because multiple `apt-get` commands cannot be run at the same time.

You could certainly use `depends` to chain these tasks together but this is
tedious and error prone. Luckily, Converge supports `groups` which makes this
much easier. We can add a named group to each task and Converge will modify the
graph so that these tasks are not run in parallel.

```hcl
task "install-tree" {
  check = "dpkg -s tree >/dev/null 2>&1"
  apply = "apt-get install -y tree"
  group = "apt"
}

task "install-jq" {
  check = "dpkg -s jq >/dev/null 2>&1"
  apply = "apt-get install -y jq"
  group = "apt"
}

task "install-build-essential" {
  check = "dpkg -s build-essential >/dev/null 2>&1"
  apply = "apt-get install -y build-essential"
  group = "apt"
}
```

And the corresponding graph:

{{< figure src="/images/dependencies/with-groups.png" caption="The graph output of the above module. The tasks in the group will not run in parallel." >}}

{{< note title="Future Improvements" >}}
In this example, we are installing packages by calling `apt-get` in Converge
tasks. We plan to build higher-level resources to handle package management that
will handle these details for you.
{{< /note >}}
