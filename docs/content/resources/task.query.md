---
title: "task.query"
slug: "task-query"
date: "2016-12-08T15:04:23-06:00"
menu:
  main:
    parent: resources
---




## Example

```hcl
task.query "hostname" {
  query = "hostname"
}

file.content "hostname-data" {
  destination = "hostname.txt"
  content     = "{{lookup `task.query.hostname.status.stdout`}}"
}

```


## Parameters

- `interpreter` (string)



- `query` (string)



- `check_flags` (list of strings)



- `exec_flags` (list of strings)



- `timeout` (optional duration)

  
Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".


- `dir` (string)



- `env` (map of string to string)





