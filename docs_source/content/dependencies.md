Using: "---
title Dependencies"
date: "2016-08-25T00:13:59-05:00"

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
Let's explore what that means with one of our graphs from before:

{{< figure src="/images/getting-started/hello-world-params.png"
           caption="A graph with a parameter. The file hello.txt depends on the name parameter." >}}

What does Converge do when you ask it to apply this graph? When Converge loads
this file, it will load then file and then start walking at the node that
doesn't have any dependencies. In this case, that's `param.name = "World"`. When
`param.name` has been successfully walked, it will move on to `File: hello.txt`.
If we're successful, the root (`/`) will be marked as successful, and our graph
will be successful. Neat!

## The Graph Command

All the graphs we've been seeing so far have just been the output of Converge's
`graph` command. When asked, Converge will load up any modules you specify and
then render them as [Graphviz](http://graphviz.org/) dot output. You can render
that like so:

```sh
$ converge graph --local yourModule.hcl | dot -Tpng > yourModule.png
```

When you're developing modules, make a habit of rendering them as graphs. It
makes it easier to think about how the graph will be executed.

## Cross-Node References

Resources may references one-another as long as the references do not introduce
circular dependencies.  When creating a reference from one node to another we
can use the `lookup` command to reference a specific field as shown in the
following example:

```hcl
param "name" {
  default="shouldrun.sample"
}

param "dest" {
  default = "{{lookup `file.content.disk.Destination`}}"
}

task "checkspace" {
  interpreter = "/bin/bash"
  check = "[[ -f {{param `name`}} ]]"
  apply = "df -h; date > {{param `name`}}"
}

task "finish checkspace" {
  check = "[[ ! -f {{param `name`}} ]];"
  apply = "rm -f {{param `name`}}"
  depends = ["task.checkspace"]
}

task "print file info" {
  interpreter = "/bin/bash"
  check = "[[ ! -f {{lookup `file.content.disk.Destination`}}]]"
  apply = "rm -f {{param `dest`}}"
}

file.content "disk" {
  destination = "diskspace.txt"
  content = "{{lookup `task.checkspace.Status.Stdout`}}"
}
```

dependencies will be automatically identified for calls to `lookup`.


`* root/task.print file info: template: check:1:10: executing "check" at <lookup `file.content...>: error calling lookup: term should be one of [Content Destination Status] not "Foo"`

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
before walking ours. They're specified as the resource type, a dot, then the
resource name. So `task "names"` above becomes `task.names`.

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
