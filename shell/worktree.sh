# Source this file from Bash or Zsh after installing the worktree binary.
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
