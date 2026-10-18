# Worktree

A small Go CLI that works entirely in Git worktrees: switch to the worktree for
a branch, and create one on the spot when it does not exist yet. It never checks
a branch out in place.
Requires Go 1.26 and Git on `PATH`.

## Installation

```sh
go install ./cmd/worktree
```

This installs the `worktree` command.

### Zsh configuration

Add this block to `~/.zshrc`. Go must already be on `PATH`.

```zsh
export PATH="${GOBIN:-$(go env GOPATH)/bin}:$PATH"

autoload -Uz compinit
compinit

eval "$(worktree init zsh)"
```

If your Zsh framework already runs `compinit`, omit those two lines and put
the `eval` line after the framework setup. Open a new shell to load the config.

### Bash configuration

Add the Go binary directory to `PATH` and load the integration in `~/.bashrc`:

```bash
export PATH="${GOBIN:-$(go env GOPATH)/bin}:$PATH"
eval "$(worktree init bash)"
```

After an update, run the install command again.

## Usage

From a Git repository or one of its worktrees:

```sh
worktree switch feature/my-change
```

`s` is a shorter alias: `worktree s feature/my-change`. The argument is a branch
name, not a worktree directory path.

The integration defines a `worktree` shell function that handles
`worktree switch <branch>` and passes other invocations to the binary. This
function is required to change the shell's directory. Without it, the binary
prints the destination path only.

If a worktree already holds the branch, the shell function changes to that
directory. The branch name must match exactly; shell completion fuzzy-matches
branches, so `worktree s feat<Tab>` offers `feature/my-change`.

If no worktree matches, the command creates one and moves you there:

- An existing local branch is checked out into the new worktree.
- A name that exists only as a remote-tracking branch creates a local branch
  that tracks the remote, through Git's own selection.
- A name that matches no branch creates a new branch, started from the branch
  checked out in the current directory.

The new worktree is created with `git worktree add` and placed at
`<main-checkout>.worktrees/<branch>/<repository>`. For a main checkout at
`/dev/worktree`, branch `feature/my-change` becomes
`/dev/worktree.worktrees/feature/my-change/worktree`. The branch path structure
is kept. The repository name is the leaf directory, so tty environments that
show the current directory keep naming the repository, not the branch. Git
rejects a branch that is already checked out in another worktree, but that
branch already has a worktree, so `switch` moves to it instead of creating one.

The binary performs any creation and prints the destination path. A child
process cannot change its parent shell's directory. Running
`command worktree switch <branch>` can create a worktree; it is not a read-only
path lookup. Completion uses local remote-tracking refs; it does not fetch.

The command works from the main checkout or a linked worktree, and supports only
standard repositories, not bare ones. Paths with spaces are supported. Failed
operations leave the current directory unchanged. Detached worktrees cannot be
selected by branch, and the command does not remove worktree directories.

Exit codes: `0` for success, `1` for an operation failure, `2` for invalid usage.

## List worktrees

List the repository's worktrees, one per line, as `<branch>\t<path>`:

```sh
worktree list
```

Detached worktrees have an empty branch field. The bare worktree of a bare
repository is not listed. This command reads only and never changes the
directory.

## Branch completion

Type `worktree switch ` and press Tab to complete a branch name.
A prefix such as `worktree switch feature/` limits the suggestions. Completion
lists sorted, unique local and remote-tracking branch names, including branches
without a worktree. Remote branches appear with and without the remote prefix.
Symbolic remote refs such as `origin/HEAD` are excluded. In a linked worktree,
completion shows the same branches; a branch with no worktree gets one created on
switch. Detached worktrees have no branch name to complete. Outside a repository,
completion stays silent.

## Structure

```text
cmd/worktree/main.go                   composition root
internal/worktree/worktree.go          entity and domain errors
internal/worktree/switch_worktree.go    switch-or-create use case and ports
internal/worktree/create_worktree.go    worktree-creation use case and outbound ports
internal/worktree/complete_branches.go  branch completion use case and inbound port
internal/worktree/cliin/               CLI inbound adapter and shell integration
internal/worktree/gitcli/              Git outbound adapter
```

The `worktree init <bash|zsh>` command prints the Bash or Zsh integration that
provides the directory change and completion. Load it with
`eval "$(worktree init <bash|zsh>)"`.

The CLI calls the `SwitchWorktree` inbound port. The use case calls the
`WorktreeLister` outbound port to resolve an existing worktree, and the
`WorktreeCreator` port (the `CreateWorktree` use case) to create one when none
matches. The Git adapter implements those ports and converts process and parsing
failures to use-case errors. Only `main` wires the adapters. The core owns the
switch-or-create rule and uses plain Go types, with no CLI or process dependencies.

Shell completion calls the private `worktree __complete switch <prefix>` CLI protocol.
The CLI calls `CompleteBranches`, which uses `WorktreeLister` and `BranchLister`.
The protocol writes one branch per line and suppresses failure diagnostics.

## Checks

```sh
go test ./...
go vet ./...
```
