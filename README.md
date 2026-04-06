# committed

[![Go reference](https://img.shields.io/badge/pkg.go.dev-jaredreisinger/committed-007D9C?logo=go&logoColor=white)](https://pkg.go.dev/github.com/jaredreisinger/committed)
![GitHub go.mod Go
version](https://img.shields.io/github/go-mod/go-version/JaredReisinger/committed?logo=go&logoColor=white)
[![GitHub Actions workflow status](https://img.shields.io/github/actions/workflow/status/JaredReisinger/committed/ci.yaml?logo=github&logoColor=white)](https://github.com/JaredReisinger/committed/actions/workflows/ci.yaml)
[![codacy
grade](https://img.shields.io/codacy/grade/f9e05f25d82e4c5d8b5421b54b49af38?logo=codacy)](https://app.codacy.com/gh/JaredReisinger/committed)
[![code coverage](https://img.shields.io/codecov/c/github/JaredReisinger/committed?logo=codecov&logoColor=white)](https://app.codecov.io/gh/JaredReisinger/committed)
![GitHub License](https://img.shields.io/github/license/JaredReisinger/committed)

A bubbletea-powered text UI that integrates as a proper `commit-msg` hook for conventional commits.

## Installation

Either grab the binary appropriate to your platform from the [Releases](https://github.com/JaredReisinger/committed/releases) page, or install directly from Go:

```bash
go install github.com/jaredreisinger/committed@latest
```

> [!NOTE]
>
> If you're using Go 1.24 or later, you can also add `committed` as a tool dependency via:
>
> ```bash
> go get -tool github.com/jaredreisinger/committed@latest
> ```


## Usage

### As a Git Hook

Install as a `commit-msg` hook:

```bash
# In your git repository
ln -s $(which committed) .git/hooks/commit-msg
```

or create a `.git/hooks/commit-msg` file that leverages it:

```bash
cat <<EOF > .git/hooks/commit-msg
#!/usr/bin/sh
committed $@
EOF

chmod +x .git/hooks/commit-msg
```

> [!NOTE]
>
> If you're using Go 1.24 or later, and added `committed` as a tool dependency, you can start it with `go tool committed`:
>
> ```bash
> cat <<EOF > .git/hooks/commit-msg
> #!/usr/bin/sh
> go tool committed $@
> EOF
>
> chmod +x .git/hooks/commit-msg
> ```


Now when you run `git commit`, the TUI will appear to help format your commit message.

### Manual Usage

You can also run the tool manually; it will load/edit the given file:

```bash
committed /path/to/COMMIT_EDITMSG
```

> [!NOTE]
>
> If you're using Go 1.24 or later, and added `committed` as a tool dependency, you can start it with::
>
> ```bash
> go tool committed /path/to/COMMIT_EDITMSG
> ```


## Configuration

The tool automatically detects conventional commit configuration from:

1. `.commitlintrc.json`
2. `.commitlintrc.{yaml,yml}`
3. `package.json` (commitizen field)

### Example `.commitlintrc.json`

```json
{
  "rules": {
    "type-enum": [2, "always", ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore", "revert"]],
    "subject-max-length": [2, "always", 72]
  }
}
```

## Features

- Interactive TUI for composing conventional commits
- (TODO) Automatic detection of project conventions
- Pre-population from existing commit messages
- (TODO) Real-time validation with helpful error messages
- Support for conventional commit format with type, scope, description, and body

## Development

This project leverages [`task`](https://taskfile.dev) to manage testing and building.

```bash
# Run tests
task test

# Build
task build

# Install locally
task install
```

Also, tool dependencies (like gcov2lcov) are now managed via the ([as of Go 1.24](https://go.dev/doc/modules/managing-dependencies#tools)) `go get -tool ...` command, so things like `go mod tidy` will automatically fetch them. Additionally, there is a `task prepare` command that functions semantically like `npm install`... it fetches dependencies and performs other one-time post-clone steps.
