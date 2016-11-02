---
title: "Resource Authors Guide"
date: "2016-09-21T10:55:55-05:00"

menu:
  main:
    parent: "converge"
    weight: 80
---

{{< warning title="Advanced Topic" >}}
This is something most people will never have to do. That said, if you know why
you're here, read on&hellip;
{{< /warning >}}

## Contributing Resources

We welcome external resource contributions, and can help you plan out how your
resource will work. If you have an idea for a new module, please open an issue
on Converge and we'll help you figure out exactly how to develop it. We'll also
gladly help you with documentation and getting everything imported in the right
places. Converge modules must currently be included in the main source of
Converge, we hope to be able to make them external at some point in the future
(at which point external contributions become even easier!)

## Check and Apply

Converge resources have a basic pattern: two operations, check and apply. This
pattern is most obvious in the shell task resource:

```hcl
task "example" {
  check = "test -f hello.txt"
  apply = "touch hello.txt"
}
```

When Converge applies this resource, the check statement is run first. If the
exit code is non-zero, Converge then runs the apply statement. Afterwards, it
runs the check statement again to see if the application was successful.

To implement something like the shell task, you would start by implementing a
struct and it's corresponding methods.
[`resource.Task`](https://godoc.org/github.com/asteris-llc/converge/resource#Task)
is the interface you'll need to implement.

```go
struct MyShellTask struct {
    resource.Status

    CheckStmt string
    ApplyStmt string
}

func (t *MyShellTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
    // your check implementation, returning t (which embeds a TaskStatus)
}

func (t *MyShellTask) Apply(context.Context) (resource.TaskStatus, error) {
    // your apply implementation, returning t (which embeds a TaskStatus)
}
```

### Task Status

Check and Apply both return
[`TaskStatus`](https://godoc.org/github.com/asteris-llc/converge/resource#TaskStatus).
You can implement that interface yourself, but the most common pattern is to
embed
[`Status`](https://godoc.org/github.com/asteris-llc/converge/resource#Status) in
the task itself. This is good for two reasons: first, this makes the fields on
your struct available for lookups. Second, and more importantly, `Status`
provides a number of helper methods to make your integration go more smoothly.

`Status` has three fields: `Differences`, `Output`, and `Level`. They all have
accurate documentation on their fields, which we will not repeat here. However,
the interaction between `Differences` and `Level` deserves special mention.
These two fields are how you control execution of your Apply method. They follow
these rules:

1. If the level is equal to
   [`resource.StatusWillChange`](https://godoc.org/github.com/asteris-llc/converge/resource#StatusLevel)
   or
   [`resource.StatusCantChange`](https://godoc.org/github.com/asteris-llc/converge/resource#StatusLevel),
   the Status will always show up as having changes.
1. Otherwise, if there are any diffs which say that they contain a difference,
   the Status will always show as having changes.

### Dealing with Errors

The default `Status` implementation has a `SetError(error)` method. When called,
it sets an appropriate error level and an internal error state that will be
presented to the user. You can also use the common Go patterm of returning `nil,
err` in your `Check` and `Apply` statements. Converge will call `SetError(err)`
automatically in this case. This gives you two options:

1. call `SetError` yourself.
1. return an error, which will be handled for you.

You shoud choose *one* of these options and do it consistently across as much of
your code as possible.

## Preparer

Before you can use your resource, it has to be deserialized from HCL. For this,
we will write a
[`resource.Resource`](https://godoc.org/github.com/asteris-llc/converge/resource#Resource).
Resource exists to render a Task's fields and return it in it's executable
state. In our case, our Resource would look like this (they're typically called
`Preparer`).

```go
type Preparer struct {
    Check string `hcl:"check"`
    Apply string `hcl:"apply"`
}

func (p *Preparer) Prepare(context.Context, resource.Renderer) (resource.Task, error) {
    return &MyShellTask{CheckStmt: check, ApplyStmt: apply}, nil
}
```

To get values other than strings (bools, ints, et cetera), you just need to
specify them. Converge will render the values and parse them from strings, if
necessary.

### Zero Values

Sometimes you need to disambiguate between a zero value the user provided and
one that Go did. In this case, use a pointer to that type. For example, your
preparer may look like this:

```go
type Preparer struct {
    Field int `hcl:"field"`
}
```

But in this case, the value of `Field` would be zero in each of the two calls
below!

```hcl
mymodule "test" {
    field = 0
}

mymodule "test" {
    # field is unspecified!
}
```

If you need to tell which case happened, use a pointer. In other words, your
preparer will now look like this:

```go
type Preparer struct {
    Field *int `hcl:"field"`
}
```

If the user provides a zero, the value will be `*0`. Otherwise, it wil be `nil`.

### Struct Tags

Other than `hcl` (which is used to specify the field name you'll accept) the
following struct tags control the values you get:

- `doc_type`: control the exact printed type in the documentation. Example:
  fields that accept
  a [duration string](https://golang.org/pkg/time/#ParseDuration) (such
  as [task.timeout]({{< ref "resources/task.md" >}})) are commonly strings
  with a `doc_type` of "duration string"

- `base`: used with numeric types to indicate a base for parsing. Does not work
  with floats. Example: [file.mode]({{< ref "resources/file.mode.md" >}})
  needs an octal number, and specifies that in this tag.

We can also do some basic validation tasks with tags:

- `required`: one valid value: `true`. If set, this field must be set in the
  HCL, but may still have a zero value (for example, `int` can still be `0`.)
  Example: [docker.container]{{< ref "resources/docker.container.md" >}} uses
  this to require an image for the container.

- `mutually_exclusive`: a comma-separated list of fields that cannot be set
  together. Example: [user]({{< ref "resources/user.user.md" >}}) uses this
  to disallow setting both `groupname` and `gid`.

- `valid_values`: a comma-separated list of values that will be accepted for
  this field.
  Example: [docker.container]({{< ref "resources/docker.container.md" >}}) uses
  this to enforce status is only `running` or `created`.

### The Renderer

The renderer is what allows your values to take input from the environment (like
calls to `param` or `lookup`.) Normally you won't need to use this, but if
you're doing something extremely custom it will be handy. If you get an error
while using the `Renderer`, return it exactly as received or wrap it with
[`errors.Wrap` or `errors.Wrapf`](https://github.com/pkg/errors). Converge uses
these signals to calculate execution order, so it needs to be able to inspect
the returned error value.

## Registering

The last thing you'll need to do is register your new resource with the loader
under it's HCL-usable name. To do so, call
[`registry.Register`](https://godoc.org/github.com/asteris-llc/converge/load/registry#Register)
with both the preparer and task, then empty-import your new module in
`load/resource.go`.

```go
func init() {
    registry.Register("mytask", (*Preparer)(nil), (*MyShellTask)(nil))
}
```
