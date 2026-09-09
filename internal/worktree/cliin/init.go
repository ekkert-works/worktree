package cliin

import (
	"fmt"
	"io"
)

// initShell writes the shell integration for the named shell to stdout.
// Users load it with eval "$(worktree init <bash|zsh>)". A binary cannot
// change its parent shell's directory, so the integration defines a worktree
// function that performs the switch and changes the directory itself.
func initShell(arguments []string, stdout, stderr io.Writer) int {
	var script string
	switch {
	case len(arguments) == 1 && arguments[0] == "bash":
		script = bashScript
	case len(arguments) == 1 && arguments[0] == "zsh":
		script = zshScript
	default:
		fmt.Fprintln(stderr, "usage: worktree init <bash|zsh>")
		return 2
	}
	if _, err := io.WriteString(stdout, script); err != nil {
		fmt.Fprintln(stderr, "worktree: cannot write shell integration")
		return 1
	}
	return 0
}

// switchFunction is shared by both shells. A child process cannot change its
// parent shell's directory, so this function performs the checkout and changes
// the shell's directory itself.
const switchFunction = `worktree() {
    if [ "$#" -ne 2 ] || [ "$1" != "switch" ]; then
        command worktree "$@"
        return $?
    fi

    local destination
    # A sentinel preserves paths that end in newlines during substitution.
    destination="$(command worktree "$@" && printf '.')" || return $?
    destination="${destination%.}"
    destination="${destination%?}"
    cd -- "$destination"
}
`

const bashScript = switchFunction + `
_worktree_complete_bash() {
    COMPREPLY=()
    if [ "$COMP_CWORD" -ne 2 ] || [ "${COMP_WORDS[1]}" != "switch" ]; then
        return 0
    fi

    local branch
    while IFS= read -r branch; do
        COMPREPLY+=("$branch")
    done < <(command worktree __complete switch "${COMP_WORDS[COMP_CWORD]}" 2>/dev/null)
}

complete -F _worktree_complete_bash worktree
`

const zshScript = switchFunction + `
_worktree_complete_zsh() {
    if [ "$CURRENT" -ne 3 ] || [ "${words[2]}" != "switch" ]; then
        return 0
    fi

    _worktree_candidates_zsh
}

_worktree_candidates_zsh() {
    local output
    local -a branches
    output="$(command worktree __complete switch "$PREFIX" 2>/dev/null)" || return 0
    [ -n "$output" ] || return 0
    branches=("${(@f)output}")
    compadd -- "${branches[@]}"
}

compdef _worktree_complete_zsh worktree
`
