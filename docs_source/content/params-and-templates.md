---
title: "Params and Templates"
date: "2016-09-30T10:34:07-05:00"

menu:
  main:
    parent: "converge"
    weight: 50
---

Almost every value in Converge can be templated. We saw the basic usage of
params in the [using dependencies]({{< ref "dependencies.md" >}}) guide, but
there's even more you can do.

## Values

First, it's helpful to realize that you can put almost any value in a param. You
can have any number, string, or boolean in a param, or a list or map of any
combination of those. (Note that currently you cannot have a list or map of
lists or maps, those may be added later.) The resource documentation for any
given module will give you a guide to the types you can use for each field. Note
that you can use the string value of any of these values as well, and Converge
will automatically convert it before providing it to the module. The conversion
rules are:

- **strings** will be used as-is after expanding all parameters and template
  blocks

- **boolean values** will be interpreted as-is if they're the literals `true` or
  `false`. If the value is a string, any capitalization of `t` or `true` will be
  truth, and any other string value will be false. Numeric values are not
  accepted as boolean values.

- Converge will parse **numbers** according to the bit size and signedness
  specified in the resource. Let's take `uint32` as an example: providing a
  negative integer or one above the ceiling for 32-bit unsigned integral values
  will result in an error. Resources also can specify a base for conversion. For
  example, [file.mode]({{< ref "resources/file.mode.md" >}})) takes an octal
  (base-8) integral value.

- **list** items will be interpreted using the semantics above

- **map** keys and values will both be interpreted using the semantics above

## Templates

Converge provides the following template functions for your use:

### Params

Use of any of these functions will create edges in the graph pointing from your
resource to the named parameter.

- **param** refers to a parameter as a single stringified value.

- **paramList** refers to a parameter as a list of values. Use this in
  combination with `range` to loop over values, as in `samples/paramList.hcl` in
  the Converge source.

- **paramMap** refers to a parameter as a map of values. Use this in combination
  with `range` to loop over values, as in `samples/paramMap.hcl` in the Converge
  source.

### Platform

- **platform** TODO

- **env** retrieves an item (named by the first argument) from an environment
  variable

### Utility

- **split** splits a string (second argument) at the instances of another string
  (first argument)

- **join** joins a list of strings (second argument) using another string (first
  argument)

- **jsonify** returns the value as a JSON string
