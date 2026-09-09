# Source this file from Bash or Zsh after installing worktree and git-wt.
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

# Git external commands cannot change the parent shell's directory.
git() {
    if [ "$#" -eq 2 ] && [ "$1" = wt ]; then
        worktree switch "$2"
        return $?
    fi
    command git "$@"
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

# Native Zsh Git completion calls _git-<subcommand>.
_git-wt() {
    [ "$CURRENT" -eq 2 ] || return 0
    _worktree_candidates_zsh
}

# Git's Bash completion (also used by some Zsh setups) calls _git_<subcommand>.
_git_wt() {
    [ "$cword" -eq 2 ] || return 0
    __gitcomp_nl "$(command worktree __complete switch "$cur" 2>/dev/null)"
}

if [ -n "${ZSH_VERSION-}" ] && command -v compdef >/dev/null 2>&1; then
    compdef _worktree_complete_zsh worktree
elif [ -n "${BASH_VERSION-}" ]; then
    complete -F _worktree_complete_bash worktree
fi
