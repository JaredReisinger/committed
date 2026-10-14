# Contributing to committed

:+1::tada: First off, thank you for taking the time to contribute! :tada::+1:

The following is a set of guidelines and expectations for contributing to committed. Note that these are _guidelines_, not set-in-stone rules. Use your best judgment, and also feel free to propose changes to this document itself in a code request, or by opening an issue.

#### Table of contents

<!-- TOC depthfrom:2 depthto:3 orderedlist:false updateonsave:true -->
- [Code of conduct](#code-of-conduct)
- [I just have a question!](#i-just-have-a-question)
- [How can I contribute?](#how-can-i-contribute)
  - [Reporting bugs](#reporting-bugs)
  - [Suggesting enhancements](#suggesting-enhancements)
  - [Pull requests / development](#pull-requests--development)
<!-- /TOC -->

## Code of conduct

This project and everyone participating in it is governed by the [code of conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## I just have a question!

If you have a question, or found what looks like a bug, please feel free to [open a new GitHub issue](https://github.com/JaredReisinger/committed/issues?q=is%3Aissue) for it. Even better, check to see if there's already been an issue that addresses the question first.

## How can I contribute?

### Reporting bugs

If you've run across what feels like a bug, first [check to see if there's already an open issue](https://github.com/JaredReisinger/committed/issues) about it. If there is, you can add any additional details you think would helpful in tracking down the issue. If there is a **closed** issue that seems to be the same thing that you're seeing, please create a new issue and reference the old one rather than re-opening the closed issue.

When describing the bug, please be as detailed as possible:

- **Use a clear and descriptive title** to help identify the problem.

- **Include the version number of committed** that you're using.

- **Provide a description or excerpt of the code** if you can.

- **Include the error message** if there was one.

- **Include a screenshot or animated GIF** if it's a rendering issue.

### Suggesting enhancements

Much as with bugs, please first [check to see if there's already an open issue](https://github.com/JaredReisinger/committed/issues) that describes the enhancement you'd like to see. If there is, you can add any additional specifics that would help the enhancement meet your use case.

When suggesting a new enhancement, please be as detailed as possible:

- **Use a clear and descriptive title** to help identify the suggestion.

- **Provide a description of the problem you're trying to solve** rather than simply suggesting the solution outright. In some cases, there might be a different approach altogether that actually fulfills your needs more directly.

- **Explain why this problem might be of general interest** and not an isolated one-off solution to your specific problem. (If you feel it's more of a one-off situation, you should still feel free to open an issue so we can discuss possible solutions!)

- **Include the version number of committed** that you're using.

### Pull requests / development

This project leverages [`task`](https://taskfile.dev), installed as a Go tool, to manage testing and building. If you _already_ have `task` installed, use `task prepare` to download all of the needed dependencies and tools, and to set `committed` up as this repo’s commit message editor.

If you _don’t_ have `task` yet, you’ll need to bootstrap yourself. Run:

```bash
go mod download
go tool task prepare   # this step can take a bit since task needs to build
```

From this point on `go tool task` will list the available task targets/recipes, such as:

```bash
# Run tests
go tool task test

# Build
go tool task build
```

Please make sure your commit messages follow the [Conventional Commits](https://www.conventionalcommits.org) standard, and use your PR to describe the specific problem you are trying to solve, or the feature you are adding and why it’s a benefit.

And thanks!
