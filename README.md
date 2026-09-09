# Worktree

A small Go CLI to switch to an existing Git worktree by branch name.
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

The sourced script defines a `git` shell function that handles `git wt <branch>`
and passes other commands to Git. Without this function, the standalone Git
subcommand prints the destination path. It cannot change the parent shell directory.

The branch must match exactly. The shell function changes the current directory.
The binary alone prints the destination path: a child process cannot change its
parent shell's directory. Use `command worktree switch <branch>` to get the path.
Paths with spaces are supported. Missing or ambiguous branches produce an error
and leave the current directory unchanged. Detached worktrees cannot be selected
by branch. This command does not create, remove, or modify worktrees.

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

Both CLI entry points call the `SwitchWorktree` inbound port. The use case calls the
`WorktreeLister` outbound port. The Git adapter implements that port and converts
process and parsing failures to use-case errors. Only `main` wires the adapters.
The core uses plain Go types and has no CLI or process dependencies.

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
