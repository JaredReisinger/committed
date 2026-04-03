package hook

// HookContext contains information about the git hook invocation.
type HookContext struct {
	MessageFilePath string // Path to the commit message file
	SourceType      string // message, template, merge, squash, or commit
	SourceObject    string // SHA1 for amend commits, empty otherwise
}
