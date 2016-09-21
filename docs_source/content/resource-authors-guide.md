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

func (t *MyShellTask) Check(resource.Renderer) (resource.TaskStatus, error) {
    // your check implementation, returning t (which embeds a TaskStatus)
}

func (t *MyShellTask) Apply() (resource.TaskStatus, error) {
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

func (p *Preparer) Render(r *resource.Renderer) (resource.Task, error) {
    check, err := render.Render("check", p.Check)
    if err != nil {
        return nil, err
    }

    apply, err := render.Render("apply", p.Apply)
    if err != nil {
        return nil, err
    }

    return &MyShellTask{CheckStmt: check, ApplyStmt: apply}, nil
}
```

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
