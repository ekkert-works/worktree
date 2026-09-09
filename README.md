# Worktree

A small Go CLI to switch to an existing Git worktree or check out a branch.
Requires Go 1.26 and Git on `PATH`.

## Installation

```sh
go install ./cmd/worktree ./cmd/git-wt
```

This installs `worktree` and the Git subcommand `git-wt`. Git finds `git-wt`
on `PATH` when you run `git wt`.

### Zsh configuration

Add this block to `~/.zshrc`. Replace the source path with your checkout path.
Go must already be on `PATH`.

```zsh
export PATH="${GOBIN:-$(go env GOPATH)/bin}:$PATH"

autoload -Uz compinit
compinit

source "$HOME/development/ekkert-works/worktree/shell/worktree.sh"
```

If your Zsh framework already runs `compinit`, omit those two lines and put
the `source` line after the framework setup. Open a new shell to load the config.

### Bash configuration

Add the Go binary directory to `PATH` and source the script in `~/.bashrc`:

```bash
export PATH="${GOBIN:-$(go env GOPATH)/bin}:$PATH"
source "$HOME/development/ekkert-works/worktree/shell/worktree.sh"
```

After an update, run the install command again and source the script again.

## Usage

From a Git repository or one of its worktrees:

```sh
git wt feature/my-change
# Equivalent command:
worktree switch feature/my-change
```

Use the branch name of an existing worktree to move to it. The argument is
a branch name, not a worktree directory path.

The sourced script defines a `git` shell function that handles `git wt <branch>`
and passes other commands to Git. This function is required to change the shell's
directory. Without it, the standalone Git subcommand prints the destination path.

If the branch has an existing worktree, the shell function changes to that
directory. The branch name must match exactly.

If the branch has no worktree, the command runs `git checkout <branch> --` in
the current directory. This is allowed only in the main checkout. In a linked
worktree, including a detached worktree, the command returns an error and keeps
the current branch. Switching to another existing worktree is still allowed.

Git applies its normal checks for local changes and remote branch selection.
An unqualified remote branch name can create a local tracking branch. A qualified
name such as `origin/feature` checks out that remote ref with a detached HEAD,
as `git checkout` does.

The binary performs the checkout, if needed, and prints the destination path.
A child process cannot change its parent shell's directory. Running
`command git wt <branch>` or `command worktree switch <branch>` can change the checked-out branch;
it is not a read-only path lookup.

Paths with spaces are supported. Failed operations leave the current directory
unchanged. Detached worktrees cannot be selected by branch. The command does not
create or remove worktree directories.

Exit codes: `0` for success, `1` for an operation failure, `2` for invalid usage.

## Branch completion

Type `worktree switch ` and press Tab to complete a branch name in Bash or Zsh.
A prefix such as `worktree switch feature/` limits the suggestions. Completion
lists sorted, unique branch names from existing worktrees. It excludes detached
worktrees and branches without a worktree. Outside a repository, it stays silent.

## Structure

```text
cmd/worktree/main.go                   composition root
cmd/git-wt/main.go                     Git subcommand composition root
internal/worktree/worktree.go          entity and domain errors
internal/worktree/switch_worktree.go    use case and inbound/outbound ports
internal/worktree/complete_branches.go  branch completion use case and inbound port
internal/worktree/cliin/               CLI inbound adapter
internal/worktree/gitcli/              Git outbound adapter
shell/worktree.sh                      Bash/Zsh directory change and completion
```

The CLI calls the `SwitchWorktree` inbound port. The use case calls the
`WorktreeLister` outbound port. The Git adapter implements that port and converts
process and parsing failures to use-case errors. Only `main` wires the adapters.
The switch use case also calls `CheckoutLocationReader` to check whether the
current directory is in a linked worktree. It calls `BranchCheckout` only when
no existing worktree matches and the current directory is in the main checkout.
The core owns this rule and uses plain Go types, with no CLI or process dependencies.

Shell completion calls the private `__complete switch <prefix>` CLI protocol.
The CLI calls `CompleteBranches`, which uses the same `WorktreeLister` port.
The protocol writes one branch per line and suppresses failure diagnostics.

## Checks

```sh
go test ./...
go vet ./...
bash -n shell/worktree.sh
zsh -n shell/worktree.sh
bash shell/completion_test.sh
zsh shell/completion_test.sh
```
