# committed

A Go-native, bubbletea-powered TUI that integrates as a proper `commit-msg` hook for conventional commits.

## Installation

```bash
go install github.com/jaredreisinger/committed@latest
```

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

Now when you run `git commit`, the TUI will appear to help format your commit message.

### Manual Usage

You can also run the tool manually; it will load/edit the given file:

```bash
committed /path/to/COMMIT_EDITMSG
```

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
- Automatic detection of project conventions
- Pre-population from existing commit messages
- Real-time validation with helpful error messages
- Support for conventional commit format with type, scope, description, and body

## Development

We leverage [`task`](https://taskfile.dev) to manage testing and building.

```bash
# Run tests
task test

# Build
task build

# Install locally
task install
```

Also, tool dependencies (like gcov2lcov) are now managed via the ([as of Go 1.24](https://go.dev/doc/modules/managing-dependencies#tools)) `go get -tool ...` command, so things like `go mod tidy` will automatically fetch them. Additionally, there is a `task prepare` command that functions semantically like `npm install`... it fetches dependencies and performs other one-time post-clone steps.
