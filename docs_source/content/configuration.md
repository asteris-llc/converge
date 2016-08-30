---
title: Configuration
date: "2016-08-29T16:51:52-05:00"
menu:
  main:
    parent: converge
    weight: 60

---

Converge sources configuration from a number of different places:

## Command-Line Flags

Command-line flags will always be considered over any other source. To view
them, send `--help` to any command. In addition, all commands have these flags:

- `--config`: set the config file (see below for more info on this file)
- `--log-level`: log level, one of `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, or
  `FATAL` (`INFO` is used by default)
- `--nocolor`: set to force colorless output

## Environment

Environment variables are the same names as command-line flags, but prefixed by
`CONVERGE_` and with dashes replaced with underscores. For example, the
`--log-level` flag can be set by setting the `CONVERGE_LOG_LEVEL` environment
variable.

## Config Files

Converge will source a single config file as a fallback. This config file can
JSON, TOML, YAML, HCL, or a Java properties file (this is detected by file
extension.) The keys of this file are the same as the command-line flags.
Converge looks in `/etc/converge/config.{ext}` by default, but you can change
this with the global `--config` flag.
