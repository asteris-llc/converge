---
title: "Getting Started"
date: "2016-08-24T23:49:44-05:00"

menu:
  main:
    parent: "converge"
    weight: 10
---

In this tutorial we're going to step through basic Converge usage by creating a
small "hello world" module. Before beginning, please make sure to
[install Converge]({{< ref "install.md" >}}). You can test your installation by
issuing `converge version`.

## Hello World!

We'll begin by writing a small module using the [file.content]({{< ref
"resources/file.content.md" >}}) resource. Put the following into
`helloWorld.hcl`:

```hcl
file.content "render" {
  destination = "hello.txt"
  content     = "Hello, World!"
}
```

## Planning

This is our first module! Let's plan out our execution first by running
`converge plan helloWorld.hcl`:

```sh
$ converge plan --local helloWorld.hcl
WARN[0000] setting session-local token                   token=7671639d-b007-4145-994b-0765080bd82c
INFO[0000] serving                                       addr=:26930 component=rpc
INFO[0000] planning                                      component=client file=hello.hcl
INFO[0000] resolving dependencies                        component=rpc function=ResolveDependencies runID=7ab3e976-8922-4df8-93ef-afd89cf0823c
INFO[0000] loading resources                             component=rpc function=SetResources runID=7ab3e976-8922-4df8-93ef-afd89cf0823c
INFO[0000] rendering                                     component=rpc function=Render runID=7ab3e976-8922-4df8-93ef-afd89cf0823c
INFO[0000] got status                                    component=client file=hello.hcl id=root/file.content.render run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root run=STARTED stage=PLAN

root/file.content.render:
        Messages:
        Has Changes: yes
        Changes:
            hello.txt: "<file-missing>" => "Hello, World!"

Summary: 0 errors, 1 changes
```

You can see our single file is going to be rendered at "hello.txt", and the file
will be created. When created, it will say "Hello, World!" Planning out your
changes is usually a good idea; all of Converge's resource types support planned
output.

## Applying

Next, let's actually make the changes, using `converge apply`:

```sh
$ converge apply --local helloWorld.hcl
WARN[0000] setting session-local token                   token=4ffe080e-bc6b-4655-b7d5-e5456beb1e67
INFO[0000] serving                                       addr=:26930 component=rpc
INFO[0000] applying                                      component=client file=hello.hcl
INFO[0000] resolving dependencies                        component=rpc function=ResolveDependencies runID=56d53bc1-6fe5-4e17-a8a0-9ba97fd0c65c
INFO[0000] loading resources                             component=rpc function=SetResources runID=56d53bc1-6fe5-4e17-a8a0-9ba97fd0c65c
INFO[0000] rendering                                     component=rpc function=Render runID=56d53bc1-6fe5-4e17-a8a0-9ba97fd0c65c
INFO[0000] got status                                    component=client file=hello.hcl id=root/file.content.render run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root/file.content.render run=STARTED stage=APPLY
INFO[0000] got status                                    component=client file=hello.hcl id=root run=STARTED stage=APPLY

root/file.content.render:
        Messages:
        Has Changes: yes
        Changes:
            hello.txt: "<file-missing>" => "Hello, World!"

Summary: 0 errors, 1 changes
```

This looks *suspiciously* similar to the output from before. That's good,
because it means Converge made exactly (and only) the changes that it planned
out. If we check by opening "hello.txt" in an editor, we'll see that it says
"Hello, World!"

## The Graph

So what's actually going on here? Converge is taking your module file and
turning it into a graph, then walking over the graph in order to make changes.
Right now we only have one resource, so the graph is pretty simple:

{{< figure src="/images/getting-started/hello-world.png"
           caption="The graph of our hello world module. Root (represented as /) depends on the file we defined." >}}

(By the way, you can find more information about this in [Dependencies]({{< ref
"dependencies.md" >}}))

## Divergence!

Converge doesn't just create resources, though: it also makes sure they stay up
to date. While you're in your editor, go ahead and change the message to
something else. I changed mine to "LOL World!" Once you've done that, run the
plan again.

```sh
$ converge plan --local helloWorld.hcl
WARN[0000] setting session-local token                   token=2157f1e9-6ca9-4e6f-a203-096608c9adcd
INFO[0000] serving                                       addr=:26930 component=rpc
INFO[0000] planning                                      component=client file=hello.hcl
INFO[0000] resolving dependencies                        component=rpc function=ResolveDependencies runID=5ca7bb15-c4b2-4453-b065-ac5cba288c82
INFO[0000] loading resources                             component=rpc function=SetResources runID=5ca7bb15-c4b2-4453-b065-ac5cba288c82
INFO[0000] rendering                                     component=rpc function=Render runID=5ca7bb15-c4b2-4453-b065-ac5cba288c82
INFO[0000] got status                                    component=client file=hello.hcl id=root/file.content.render run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root run=STARTED stage=PLAN

root/file.content.render:
        Messages:
        Has Changes: yes
        Changes:
            hello.txt: "LOL World!" => "Hello, World!"

Summary: 0 errors, 1 changes
```

We see that we're going to change our message back to what's specified in the
module. Handy! A quick `converge apply helloWorld.hcl` and we're back to normal.

## Params

Now let's add the ability to greet someone in particular, instead of the whole
world. We're going to use [params]({{< ref "resources/param.md" >}}) for this.
Change your `helloWorld.hcl` to look like this:

```hcl
param "name" {
  default = "World"
}

file.content "render" {
  destination = "hello.txt"
  content     = "Hello, {{param `name`}}!"
}
```

This [param]({{< ref "resources/param.md" >}}) allows us to add a parameter to
our module when we call it. (Notice that we're using the result of our parameter
in a template block in the `content` stanza of `file.content.render`.)

{{< note title="Templates" >}}
Converge uses Go's `text/template` library. You can template most stanzas in
your resources. The
[Go Documentation on `text/template`](https://golang.org/pkg/text/template/) is
a handy reference to keep around as you're finding your feet with these
templates.
{{< /note >}}

Let's change the name in the template to your name (I'm going to assume it's
"Spartacus".) We'll use the `-p` flag to `converge plan` to see what'll happen:

```sh
$ converge plan --local -p name=Spartacus content/helloWorld.hcl
WARN[0000] setting session-local token                   token=dc73ca2e-d5dc-46d6-a2e4-81b848e7309c
INFO[0000] serving                                       addr=:26930 component=rpc
INFO[0000] planning                                      component=client file=hello.hcl
INFO[0000] resolving dependencies                        component=rpc function=ResolveDependencies runID=91562f11-df7b-4e8e-8fda-d341531afe92
INFO[0000] loading resources                             component=rpc function=SetResources runID=91562f11-df7b-4e8e-8fda-d341531afe92
INFO[0000] rendering                                     component=rpc function=Render runID=91562f11-df7b-4e8e-8fda-d341531afe92
INFO[0000] got status                                    component=client file=hello.hcl id=root/param.name run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root/file.content.render run=STARTED stage=PLAN
INFO[0000] got status                                    component=client file=hello.hcl id=root run=STARTED stage=PLAN

root/file.content.render:
        Messages:
        Has Changes: yes
        Changes:
            hello.txt: "LOL World!" => "Hello, Spartacus!"

Summary: 0 errors, 1 changes
```

Makes sense, right? When we provide the param, it's value is used instead of the
default of "World".

By the way, how does this effect our graph? Well, we've added a new resource.
Normally, you'd have to [explicitly specify dependencies]({{< ref
"dependencies.md" >}}), but Converge will look inside our template strings for
references to Params. That means our graph is all hooked up, and looks like
this:

{{< figure src="/images/getting-started/hello-world-params.png"
           caption="Our graph with parameter. The file now depends on the name parameter." >}}

## Modules Calling Modules

This is all well and good, but we don't want to have to write things the same
*every time* right? Well, good news: that's what modules are for! Now that
you've written a module, you can require it from any other module to add it to
your tree. Create a new module, let's call it `helloYou.hcl`:

```hcl
module "helloWorld.hcl" "hello" {
  params {
    name = "Spartacus"
  }
}
```

Now try running `converge plan helloYou.hcl`. The same thing happens as if you
had called the module yourself!

But once again, how does this effect our graph? You remember before that we had
a root, a file resource, and a param resource. We still have all those things,
but with a wrinkle: we've seen `/` as our root module in all the previous
diagrams, but `/` is really just an alias for the root. Since we're not
requiring `helloWorld.hcl` from `helloYou.hcl`, Converge will reason about it as
a separate module. But really, these are all part of the same tree, and Converge
will walk over them as nodes.

{{< figure src="/images/getting-started/hello-you.png"
           caption="Our graph, but with our original module as a dependent module." >}}

## What's Next?

A great next step is to try and make something simple with Converge! Try
installing something from Brew, if you're running on OSX, or symlinking a
dotfile into place.

If you're ready, you can
[read about how all those dependencies *actually work*]({{< ref
"dependencies.md" >}})!
