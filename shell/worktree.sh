# Source this file from Bash or Zsh after installing worktree.
# A child process cannot change its parent shell's directory, so this
# function performs the checkout and changes the shell's directory itself.
worktree() {
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

if [ -n "${ZSH_VERSION-}" ] && command -v compdef >/dev/null 2>&1; then
    compdef _worktree_complete_zsh worktree
elif [ -n "${BASH_VERSION-}" ]; then
    complete -F _worktree_complete_bash worktree
fi
