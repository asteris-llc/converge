# Contribution Guide

Hi! Thanks for being interested in contributing to Converge. We're really happy
to have you (we have mini-parties in Slack whenever someone external opens a
PR!)

Here's a quick guide for how you should expect the contribution process to go.
Above all, the core team and everyone involved with Converge is expected to
follow the [code of conduct](CODE_OF_CONDUCT.md).

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Contribution Guide](#contribution-guide)
    - [Filing Bugs and Getting Help](#filing-bugs-and-getting-help)
    - [Feature Requests](#feature-requests)
        - [Contributing Features and Fixes](#contributing-features-and-fixes)
            - [Documentation](#documentation)
    - [Code of Conduct](#code-of-conduct)

<!-- markdown-toc end -->

## Filing Bugs and Getting Help

Have you found an issue with Converge? Boo, but thanks! Issues are free, so open
one. It'd help us out if you'd search for your issue before opening a new one,
but if in doubt just go ahead and open one. We'll merge them as necessary.

When you're opening an issue, it'll save everyone some time if you could attach
the output of `converge version` and the output of the command you're having
trouble with at the DEBUG level (set `--log-level=debug` and censor if
necessary). If you don't attach that info our first comment will probably be
asking for it, so doing that will speed up the process of getting you unblocked.

Once you open an issue, it can help to also mention your issue in
the [Converge Slack channel](http://converge-slack.aster.is). This will let us
talk back and forth and figure things out even faster.

## Feature Requests

We really want to hear how you want to use Converge, and what you think it could
do for you. That said, we often have a pretty good idea of the features on our
roadmap. If Converge doesn't do something you want, open an issue and we'll
figure out how it can fit into the overall picture.

### Contributing Features and Fixes

If you just want to contribute a new feature yourself: wow, thanks! But please
still open an issue before you begin work so we can make sure that we're not
already working on something similar.

Bug fixes are also especially welcome, probably even more than new features! If
you find something wrong and easily correctable in our code, a small diff is
often much easier to reason about than describing a problem and solution.

That said, we have a number of bars for any contribution to clear:

- Your source code must be gofmt'd
- It must pass integration testing (see below)
- If it introduces a new feature or changes an existing feature, that feature
  must be documented (how else will people find out about your awesome work?)
- It can't change existing syntax except in extremely well-reasoned cases. If
  you're changing the syntax of existing features, definitely open an issue
  first so we can discuss.

### Automated Testing

We want to make sure that Converge is solid software, so to do that, we have
various flavors of automated tests, and they are all encoded in our `Makefile`.
Here are the ones that you should use when adding features:

- `make test`: Unit tests. This includes `go test`, blackbox testing, and even
   testing and formatting each converge file in `samples`
- `make lint`: Lint and vet go code
- `make bench`: Benchmark tests. Use this for analyzing performance
- `make integration`: Integration tests. There are two varieties here:
    - quick integration: This runs docker containers with the cross-compiled
      ELF binary as Converge servers and sends commands via gRPC
    - full integration: This activates only when you have
      `CVG_FULL_INTEGRATION` set in your environment, but it's currently a NOP

#### Documentation

When contributing documentation, please just change the markdown files in
`docs_source`. If your PR includes the changed HTML under `docs` it's much
harder to review. Pretty please?

## Code of Conduct

Participation in the development process of Converge is subject to
the [code of conduct](CODE_OF_CONDUCT.md). Please familiarise yourself with that
document. If nothing else, it will let you know what to do if something goes
terribly wrong. But aside from that, the code has positive standards of how
community members should behave in their interactions with others.
