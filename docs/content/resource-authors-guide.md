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
return
[`Status`](https://godoc.org/github.com/asteris-llc/converge/resource#Status).

`Status` provides a number of helper methods to make your integration go more
smoothly.  Of particular note are:

1. [`RaiseLevel`](https://godoc.org/github.com/asteris-llc/converge/resource#Status.RaiseLevel)
   which allows you to increase the level of the error
1. [`AddMessage`](https://godoc.org/github.com/asteris-llc/converge/resource#Status.AddMessage)
   which allows you to add a message that will be displayed to the user
1. [`AddDifference`](https://godoc.org/github.com/asteris-llc/converge/resource#Status.AddDifference)
   which inserts a difference that will be displayed to the user

`Status` has three fields: `Differences`, `Output`, and `Level`. They all have
accurate documentation on their fields, which we will not repeat here. However,
the interaction between `Differences` and `Level` deserves special mention.
These two fields are how you control execution of your Apply method. They follow
these rules:

1. If the level is equal to
   [`resource.StatusWillChange`](https://godoc.org/github.com/asteris-llc/converge/resource#StatusLevel)
   or
   [`resource.StatusCantChange`](https://godoc.org/github.com/asteris-llc/converge/resource#StatusLevel),
   the Status will always show up as having changes
1. Otherwise, if there are any diffs which say that they contain a difference,
   the Status will always show as having changes

### Dealing with Errors

The default `Status` implementation has a `SetError(error)` method. When called,
it sets an appropriate error level and an internal error state that will be
presented to the user. You can also use the common Go patterm of returning `nil,
err` in your `Check` and `Apply` statements. Converge will call `SetError(err)`
automatically in this case. This gives you two options:

1. call `SetError` yourself
1. return an error, which will be handled for you

You shoud choose *one* of these options and do it consistently across as much of
your code as possible.

## Task

The
[`resource.Task`](https://godoc.org/github.com/asteris-llc/converge/resource#Task) interface
is what you will implement to have converge run your `Check` and `Apply`
methods.  The `export` and `re-export-as` tags in your `resource.Task`
implementation are used to define lookup methods from within converge.

Using the `Shell` module as an example we can see how tasks should be
implemented:

```go
type Shell struct {
    CmdGenerator   CommandExecutor
    CheckStmt      string                 `export:"check"`
    ApplyStmt      string                 `export:"apply"`
    Dir            string                 `export:"dir"`
    Env            []string               `export:"env"`
    Status         *CommandResults        `re-export-as:"status"`
    CheckStatus    *CommandResults        `export:"checkstatus"`
    HealthStatus   *resource.HealthStatus `export:"healthstatus"`
    renderer       resource.Renderer
    ctx            context.Context
    exportedFields resource.FieldMap
}
```

### Exporting Values

Converge will automatically extract values from a `resource.Task` that are
annotated with the `export` or `re-export-as` struct tags.  For fields that are
exported with `export`, they can be referenced directly.  For example if we have
the following task which is implemented with `Shell`:

```hcl
task "foo" {
  check = "test -f foo.txt"
  apply = "touch foo.txt"
}
```

We may reference any of the fields exported by `Shell` in a lookup by typing
`"{{lookup 'task.foo.<field name>'}}"` where `<field name>` is any exported
fields; for example `"{{lookup 'task.foo.dir'}}"` or `"{{lookup
'task.foo.env'}}"`.  Re-exported fields will provide a namespace for structs
that also export values.  In our `Shell` example we are re-exporting a
`CommandResults` struct:

```go
type CommandResults struct {
    ResultsContext
    ExitStatus uint32 `export:"exitstatus"`
    Stdout     string `export:"stdout"`
    Stderr     string `export:"stderr"`
    Stdin      string `export:"stdin"`
    State      *os.ProcessState
}
```

Because `CommandResults` exports `stdout`, `stderr`, and `stdin`, and has been
re-exported by the `Shell` module as `status`, we can reference these values
under `status`, for example: `"{{lookup 'task.foo.status.stdout'}}"` or
`"{{lookup 'task.foo.status.stderr'}}"`.

Below is a complete example of using a lookup to reference the exported and
re-exported fields from the `Shell` module:

```hcl
task "echo" {
  check = "test -f example.txt"
  apply = "echo 'executing script' | tee example.txt"
}

file.content "task-results" {
  destination = "results.txt"
  content = "{{lookup `task.echo.check`}}; {{lookup `task.echo.apply`}} -> {{lookup `task.echo.status.stdout`}}"
}
```

This example shows how we can reference specific exported fields such as `check`
and `apply`, and also the re-exported fields from our `status`.

#### Semantics of Exported Fields

1. Fields that are tagged with `export` will be exported
1. Named structs that are tagged with `export` will be exported as a struct
1. Embedded structs will have their exported fields exported in the namespace of
   the containing struct
1. Embedded interfaces will not be exported, nor have their fields exported
1. If an embedded struct field name collides with a field from the struct it's
   embedded in, both will be exported with the embedded struct being accessible
   with 'StructName.FieldName'
1. Fields exported with `re-export-as` must be structs or pointers to structs
1. Structs exported with `re-export-as` will have their exported elements
   available under the name that the struct is re-exported as

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

- `base`: used with numeric types to indicate a base for parsing. Does not work
  with floats. Example: [file.mode]({{< ref "resources/file.mode.md" >}})
  needs an octal number, and specifies that in this tag.

We can also do some basic validation tasks with tags:

- `required`: one valid value: `true`. If set, this field must be set in the
  HCL, but may still have a zero value (for example, `int` can still be `0`.)
  Example: [docker.container]({{< ref "resources/docker.container.md" >}}) uses
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
