// Committed is a conventional-commit authoring tool. It is a bubbletea-based
// text UI in the vein of [commitizen] or [git-cz], intended to make it easier
// to write a conventional commit message correctly, and harder to get wrong.
//
// Unlike many other commit-helper tools, it is not expected to be invoked
// directly on the command-line; it is expected to be invoked as a Git
// [commit-msg] hook. This means that it does not create the commit in and of
// itself, it is only concerned with setting up the commit message.
//
// Committed attempts to discover any conventional-commit configurations in use
// by the current repo to ensure it provides an accurate reflection of the
// expected types, description length requirements, body format, and so on.  In
// theory, any message created by committed should necessarily pass a commitlint
// check.
//
// [commitizen]: https://commitizen-tools.github.io/commitizen/
// [git-cz]: https://github.com/streamich/git-cz
// [commit-msg]: https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks
package main

import (
	"github.com/jaredreisinger/committed/cmd"
)

func main() {
	cmd.Execute()
}
