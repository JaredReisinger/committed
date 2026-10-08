# To hook or not to hook

> **TL; DR** Don’t hook

There’s a long history of `git-cz` and `commitizen/cz-cli` and other similar tools using something like `husky` and the `prepare-commit-msg` hook to provide behavior so that “git commit” in a given repo can “magically” invoke a conventional-commit-aware editor.  To a large extent, the user experience is perfect... you don't need to learn any new behaviors at all; just `git commit` like always and something smart helps you write your commit message.

But it turns out that the idea of a hook-based solution is really anathema to Git’s design, and the choice to use hooks was, in hindsight, problematic. To play nicely with Git’s design, there are really only two options:

1. **Provide a completely out-of-band command that invokes commit internally.** This is basically what `git cz` or `cz` do: they manage then _entire_ commit process themselves, ultimately using something like `git commit -m` to prevent the default editor from starting.

2. **Provide a `GIT_EDITOR`/`core.editor` that can be invoked.** This allows for `git commit` to again Just Work™, but comes with the caveat that the editor will be invoked for more than just basic commits, and won’t really get the extra context that `prepare-commit-msg` gets.

After much deliberation, I’m leaning towards recommending the second option as the solution with `committed`, because the vast majority of `core.editor` needs are at least commit-adjacent, and for those that aren’t, a very basic editor will suffice.  Specifically, using `core.editor` means it will be invoked with:

- **`.git/COMMIT_EDITMSG`** for commits, amends, cherry-pick, and rebase reword
- **`.git/MERGE_MSG`** for merge
- **`.git/SQUASH_MSG`** for squash
- **`.git/TAG_EDITMSG`** for tag
- **`.git/git-rebase-todo`** for rebase picker (should warn)

The first three are really variations on commits, tags just need a very simple editor, and rebase picking is okay with a simple editor _and_ we can detect and inform the user that they can use the `sequence.editor` config setting if they want to use a different editor.
