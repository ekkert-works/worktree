# Source this file from Bash or Zsh after installing git-wt.
# Git external commands cannot change the parent shell's directory.
git() {
    if [ "$#" -ne 2 ] || [ "$1" != wt ]; then
        command git "$@"
        return $?
    fi

    local destination
    # A sentinel preserves paths that end in newlines during substitution.
    destination="$(command git wt "$2" && printf '.')" || return $?
    destination="${destination%.}"
    destination="${destination%?}"
    cd -- "$destination"
}

_worktree_candidates_zsh() {
    local output
    local -a branches
    output="$(command git-wt __complete switch "$PREFIX" 2>/dev/null)" || return 0
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
    __gitcomp_nl "$(command git-wt __complete switch "$cur" 2>/dev/null)"
}
