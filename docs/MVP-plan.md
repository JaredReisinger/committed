# Go Conventional Commit TUI Hook Tool — MVP Plan

> [!NOTE]
>
> This is a record of the Copilot plan for starting this project.
>
> Apparently, I picked the wrong hook... we should be using `commit-msg`, not `prepare-commit-msg`.

## Overview

Build a go-native, bubbletea-powered TUI that integrates as a `prepare-commit-msg` hook. It detects project conventions via config files, collects commit type/summary/details in a 3-field form with smart validation, parses existing messages when present, and writes the formatted message back to the commit file without performing the commit itself.

---

## User Requirements & Decisions

### Workflow Integration
- **Primary use case**: Invoked as the `prepare-commit-msg` hook
- **Optional enhancement**: `--commit` CLI flag to perform the actual commit (future)

### Configuration Support
- **Start with**: `.commitlintrc.json` and `.commitlintrc.yaml`
- **Also support**: `commitizen` field from `package.json`
- **Skip**: JavaScript config files (`.commitlintrc.js`) to avoid runtime evaluation
- **Design**: Extensible config loader for adding support for additional formats later

### TUI Field Strategy
- **Type field**: Selector (dropdown-style navigation through configured types)
- **Summary field**: Text input with max-length validation (from config)
- **Details field**: Multi-line text input with auto-wrapping (no user concern for line lengths)

### Existing Message Handling
- **If parsable as conventional commit**: Decompose into type/scope/description/body/footer and populate fields
- **If not parsable**: Treat entire message as details field content
- **Allow user to edit**: All fields are editable before submission

### Project Structure
- Standard Go CLI: `cmd/` for CLI entry points, `pkg/` for public packages, `internal/` for private packages
- Bubbletea patterns: To be determined during Phase 5; follow bubbletea conventions

### Key Constraint
- **No auto-commit**: Tool only edits the commit message file; git continues its normal flow

---

## Implementation Plan

### Phase 1: Project Setup & CLI Foundation

**Goal**: Initialize project structure and Cobra CLI with hook subcommand

**Tasks**:
1. Initialize Go module: `go mod init github.com/jaredreisinger/committed`
2. Create project structure:
   - `main.go` — Entry point
   - `cmd/root.go` — Cobra root command
   - `cmd/hook.go` — Cobra hook subcommand
   - `pkg/` — Public packages (tui, commit)
   - `internal/` — Private packages (config, hook)
3. Set up Cobra CLI with root command:
   - `committed` (root) — Display help
   - `committed hook` — Prepare-commit-msg hook implementation
   - Global flags: `--version`, `--help`
4. Add basic error handling and logging infrastructure
5. Add bubbletea v2 dependency: `go get charm.sh/bubbletea@latest`
6. Verify clean build

**Deliverable**: Compilable Go binary with CLI structure ready for feature implementation

**Parallel**: Can start immediately

---

### Phase 2: Config File Detection & Parsing

**Goal**: Auto-detect and load conventional-commits configuration from project files

**Tasks**:
1. Create `internal/config/types.go`:
   - `Config` struct with fields: `Types []string`, `Scopes []string`, `SubjectMaxLength int`, `BodyMaxLineLength int`, `HeaderMaxLength int`, etc.
   - Default values for all fields (sensible fallback for vanilla conventional commits)

2. Create `internal/config/loader.go`:
   - Function to search and load config files in priority order:
     1. `.commitlintrc.json`
     2. `.commitlintrc.yaml` / `.commitlintrc.yml`
     3. `package.json` (extract `commitizen` field)
   - Parse JSON config (commitlint standard format)
   - Parse YAML config (commitlint standard format)
   - Extract commitizen config from package.json
   - Merge with defaults if fields missing
   - Return Config struct or error

3. Create `internal/config/defaults.go`:
   - Provide sensible defaults (e.g., `Types: ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore", "revert"]`)
   - Provide default line length limits (e.g., subject max 72 chars, body max 100 chars)

4. Implement config validation:
   - Ensure `types` array is non-empty
   - Validate line length constraints are reasonable
   - Warn if optional fields are missing but expected

5. Create unit tests:
   - Test loading from each config file format
   - Test default values apply when config missing
   - Test error handling for malformed config
   - Test config merging (partial config + defaults)

**Deliverable**: Robust config loader that can be called from hook handler

**Parallel**: Can start after Phase 1

---

### Phase 3: Conventional Commit Parsing

**Goal**: Parse commit messages into structured form for TUI population

**Tasks**:
1. Evaluate and integrate Go conventional commit parser library:
   - Recommend: `github.com/conventionalcommit/parser`
   - Fallback: `github.com/leodido/go-conventionalcommits` or similar
   - Create `pkg/commit/types.go` with exposed types:
     ```go
     type Message struct {
         Type        string
         Scope       string (optional)
         Description string
         Body        string
         Footers     []Footer
         RawMessage  string
     }
     type Footer struct {
         Token string
         Value string
     }
     ```

2. Create `pkg/commit/parser.go`:
   - Function `ParseMessage(raw string) (*Message, error)`
   - Handle partial/malformed messages gracefully:
     - Extract what can be parsed
     - Return non-nil Message with populated fields
     - Return error if truly unparsable
   - Function `Format(msg *Message) string` to convert back to conventional format

3. Create comprehensive unit tests:
   - Valid conventional commits (various formats)
   - Commits with footers (Signed-off-by, Closes, etc.)
   - Breaking changes (!) marker
   - Partial messages (only type, no scope, etc.)
   - Completely malformed input
   - Empty input

**Deliverable**: Reliable commit message parser for use in TUI + hook

**Parallel**: Can start after Phase 1 (independent of Phase 2)

---

### Phase 4: File I/O for Hook Integration

**Goal**: Handle reading/writing commit message files per git hook contract

**Tasks**:
1. Create `internal/hook/types.go`:
   - `HookContext` struct with fields: `MessageFilePath string`, `SourceType string` (message/template/merge/squash/commit), `SourceObject string` (optional commit hash for -c/-C/--amend)

2. Create `internal/hook/handler.go`:
   - Function `ParseHookArgs(os.Args) (*HookContext, error)` to extract git hook parameters
   - Function `ReadMessageFile(filepath string) (string, error)` to read existing commit message
   - Wrapper to handle both cases: hook invocation and manual testing

3. Create `internal/hook/writer.go`:
   - Function `WriteMessageFile(filepath, content string) error` to write formatted commit message back
   - Preserve file mode/permissions if possible
   - Add error handling for write failures (e.g., full disk, permission denied)

4. Create unit tests:
   - Mock file operations
   - Test reading various message file contents
   - Test writing and verifying file contents
   - Test error cases (missing file, read-only, etc.)

**Deliverable**: Safe file I/O layer for hook message handling

**Parallel**: Can start after Phase 1 (independent of Phases 2-3)

---

### Phase 5: Bubbletea TUI Implementation

**Goal**: Build interactive 3-field form for conventional commit composition

**Tasks**:
1. Create `pkg/tui/model.go`:
   - Define bubbletea `Model` struct containing:
     - `fields`: Array of 3 fields (Type, Summary, Details)
     - `focusedField`: int (0-2, tracks which field has focus)
     - `config`: Config reference (for validation rules)
     - `existingMessage`: Parsed message (for pre-population)
     - Field states: value, cursor position, validation errors

2. Implement field models:
   - **Type field**: Selector with dropdown behavior
     - Display current type selection
     - Arrow keys (↑/↓) to navigate enum values from config
     - ENTER to confirm
   - **Summary field**: Text input with validation
     - Display current text with cursor
     - Show character count and max-length
     - Show validation error if exceeds max-length (red highlight)
     - TAB/SHIFT+TAB to move to next field
   - **Details field**: Multi-line editor
     - Display wrapped text
     - Multi-line input editing
     - Cursor navigation (arrows, end-of-line, etc.)
     - TAB/SHIFT+TAB for field nav

3. Implement bubbletea interface methods:
   - `Update(msg tea.Msg) (tea.Model, tea.Cmd)` — Handle keyboard input and state transitions
   - `View() string` — Render the UI with clear field separation, labels, and validation feedback

4. Implement keyboard navigation:
   - TAB / SHIFT+TAB — Move between fields
   - ENTER in Type field — Confirm selection, move to Summary
   - ENTER in Summary field — Move to Details (allow multi-line in Details field)
   - ESCAPE / Ctrl+C — Cancel, return error
   - Ctrl+D — Submit form

5. Implement field validation:
   - Type: Must be in config.Types enum
   - Summary: Must not exceed config.SubjectMaxLength
   - Details: No validation (auto-wrap)

6. Create pre-population logic:
   - Accept existing parsed message on startup
   - Populate Type, Summary (description), Details (body) from parsed message if available
   - If message is malformed, put entire content in Details

7. Implement output function:
   - `RunTUI(ctx context.Context, config *Config, existingMessage *Message) (*Message, error)`
   - Returns structured Message with user-entered type, summary, details
   - Returns error if user cancelled or validation failed

8. Manual testing:
   - Test with various terminal sizes
   - Test all keyboard combinations
   - Test pre-population with different message formats
   - Test validation feedback

**Deliverable**: Functional bubbletea TUI for commit composition

**Depends on**: Phase 2 (config for validation rules) and Phase 3 (parser for pre-population)

---

### Phase 6: Integration & Hook Wiring

**Goal**: Wire all components together and handle prepare-commit-msg hook contract

**Tasks**:
1. In `cmd/hook.go`, implement the hook subcommand:
   - Parse hook parameters (filename, source type) via `internal/hook/handler.go`
   - Load project config via `internal/config/loader.go`
   - Read existing message from file via `internal/hook/handler.go`
   - Parse existing message via `pkg/commit/parser.go`
   - Launch TUI via `pkg/tui` with config and parsed message
   - Format final message (reconstruct from TUI fields) via `pkg/commit/parser.go`
   - Write message back to file via `internal/hook/writer.go`
   - Exit with code 0 on success, non-zero on failure

2. Implement error handling:
   - Config not found → Use defaults, continue
   - Message file not readable → Exit non-zero
   - TUI cancelled by user (^C) → Exit non-zero
   - TUI validation error → Show error, re-prompt or exit
   - Message write failure → Exit non-zero with error message

3. Create end-to-end integration tests:
   - Simulate git hook call with various commit scenarios
   - Verify output message format
   - Verify file is updated correctly
   - Verify exit codes

4. Manual testing with git:
   - Install binary as actual prepare-commit-msg hook
   - Run `git commit -m` and verify TUI appears
   - Complete commit flow and verify message is correct
   - Test with merge commits, squash commits

**Deliverable**: Fully functional prepare-commit-msg hook

**Depends on**: All previous phases

---

### Phase 7: Testing & Refinement

**Goal**: Comprehensive testing, documentation, and edge case handling

**Tasks**:
1. Integration tests across all phases
2. Edge case testing:
   - Empty repo, no config files
   - Merge commits (source type: merge)
   - Squash commits (source type: squash)
   - Amend commits (source type: commit + object)
   - Existing messages with special characters or unicode
   - Very long subject lines
   - Terminal size edge cases (very narrow, very small)
   - Missing permissions (read-only commit file)

3. Error messages and user feedback:
   - All error messages clear and actionable
   - Validation errors show helpful hints (e.g., "Subject must be ≤ 72 chars")
   - Recovery instructions (e.g., "Press Ctrl+D to submit, Esc to cancel")

4. Documentation:
   - README with installation instructions
   - Hook setup guide (how to install as prepare-commit-msg hook)
   - Configuration examples (.commitlintrc.json, .commitlintrc.yaml)
   - Troubleshooting section

5. CLI help:
   - `committed --help` shows overview
   - `committed hook --help` shows hook-specific options

**Deliverable**: Production-ready tool with comprehensive testing

**Depends on**: Phase 6

---

## Verification Checkpoints

| Checkpoint | How to Verify | Success Criteria |
|------------|---------------|------------------|
| Config loading | Run with sample `.commitlintrc.json` and `.commitlintrc.yaml` | Tool loads config correctly, uses defaults when missing |
| Commit parsing | Feed various conventional commit formats through parser | Type, scope, description extracted correctly |
| TUI navigation | Manually enter all three fields | Tab/arrow keys work, validation shows on summary, no crashes |
| TUI population | Pre-load existing conventional commit message | Fields populate with parsed values, user can edit |
| Hook integration | Install as git hook, run `git commit` | TUI appears, commit message saved correctly, exit code 0 |
| Edge cases | Test empty repo, merge commits, malformed messages | Tool handles gracefully without crashing |
| File I/O | Verify message file is updated | Message file contains formatted output after TUI |

---

## Critical Files to Create

```
cmd/
├── committed/
│   ├── main.go              # Entry point
│   └── hook.go              # Hook subcommand implementation

pkg/
├── commit/
│   ├── types.go             # Message, Footer types
│   └── parser.go            # ParseMessage, Format functions
└── tui/
    ├── model.go             # Bubbletea Model, Update, View, RunTUI
    └── (optional) styles.go # UI styling constants

internal/
├── config/
│   ├── types.go             # Config struct
│   ├── loader.go            # LoadConfig, file detection
│   └── defaults.go          # Default values
└── hook/
    ├── types.go             # HookContext
    ├── handler.go           # ParseHookArgs, ReadMessageFile
    └── writer.go            # WriteMessageFile

docs/
└── MVP-plan.md              # This file

go.mod                        # Module definition
go.sum                        # Dependency lock
```

---

## Dependencies (go.mod)

```
github.com/spf13/cobra       # CLI framework
charm.sh/bubbletea           # TUI library (v2+)
sigs.k8s.io/yaml            # YAML parsing
github.com/conventionalcommit/parser  # Commit parsing (or alternative)
```

---

## Architecture & Design Decisions

### Config Hierarchy
- Detect in priority order: `.commitlintrc.json` → `.commitlintrc.yaml` → `package.json`
- Skip JavaScript config (`.commitlintrc.js`) to avoid runtime evaluation
- Use sensible defaults for all unspecified fields

### Parser Library
- Integrate existing Go conventional-commit parser rather than building custom
- Allows for community maintenance and compatibility with spec

### TUI Field Strategy
- Three explicit fields (Type, Summary, Details) rather than free-form text
- Guides user through conventional commit structure
- Type field uses selector (dropdown) to enforce enum values from config
- Summary field validates max-length in real-time
- Details field uses auto-wrapping to avoid user line-length concerns

### Scope Field
- MVP focuses on Type/Summary/Details
- Scope support can be added in future phase if config defines scope-enum
- Design does not preclude adding scope field later

### No Commit Execution
- Tool only edits the commit message file
- Git's `prepare-commit-msg` hook runs before user editor, so our tool stages message and returns control to git
- User then has opportunity to review in `$EDITOR` before final commit/abort

---

## Future Enhancements (Post-MVP)

- Support scope field selector if config.scope-enum defined
- Support footer selection (e.g., Closes #123, Signed-off-by)
- Support breaking-change marking (!)
- Support commitlint `.config.js` evaluation (requires careful sandboxing)
- Add `--commit` flag to perform git commit automatically after TUI
- Add template-based message generation
- Add i18n for non-English UIs
- Performance: Cache config detection across multiple hook calls

---

## Notes for Implementation

1. **Bubbletea Version**: User specified `charm.land/bubbletea/v2`. Verify v2 API stability before heavy development.
2. **Keyboard Handling**: Ensure Ctrl+C and Ctrl+D are handled consistently (cancel vs. submit).
3. **File Encoding**: Ensure commit message files are read/written as UTF-8 (git's default).
4. **Error Recovery**: If TUI crashes, gracefully return error and exit non-zero without modifying commit file.
5. **Testing Strategy**: Mock file I/O and config loading to avoid test side effects. Use integration tests sparingly for critical paths.
