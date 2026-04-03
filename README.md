# committed

A Go-native, bubbletea-powered TUI that integrates as a proper prepare-commit-msg hook for conventional commits.

## Installation

```bash
go install github.com/jaredreisinger/committed/cmd/committed@latest
```

## Usage

### As a Git Hook

Install as a `prepare-commit-msg` hook:

```bash
# In your git repository
ln -s $(which committed) .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg
```

Now when you run `git commit`, the TUI will appear to help format your commit message.

### Manual Usage

You can also run the tool manually:

```bash
committed hook /path/to/COMMIT_EDITMSG
```

## Configuration

The tool automatically detects conventional commit configuration from:

1. `.commitlintrc.json`
2. `.commitlintrc.yaml` / `.commitlintrc.yml`
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

```bash
# Run tests
go test ./...

# Build
go build -o committed ./cmd/committed

# Install locally
go install ./cmd/committed
```