# Run with both Bash and Zsh from the repository root.
set -e

project_directory="$PWD"
scratch_directory=$(mktemp -d /tmp/worktree-completion.XXXXXX)
trap 'rm -rf -- "$scratch_directory"' EXIT
go build -o "$scratch_directory/bin/worktree" ./cmd/worktree
export PATH="$scratch_directory/bin:$PATH"

if [ -n "${ZSH_VERSION-}" ]; then
    autoload -Uz compinit
    compinit -u -D
fi
source "$project_directory/shell/worktree.sh"

git init -q --initial-branch=main "$scratch_directory/repository"
cd "$scratch_directory/repository"
git -c user.name=Test -c user.email=test@example.com -c commit.gpgsign=false commit -q --allow-empty -m initial
git worktree add -q -b feature/one "$scratch_directory/one tree"
git worktree add -q -b feature/two "$scratch_directory/two tree"
git worktree add -q --detach "$scratch_directory/detached"
git branch unused

if [ -n "${ZSH_VERSION-}" ]; then
    [ "${_comps[worktree]}" = _worktree_complete_zsh ]
    # Capture candidates at the completion boundary without an interactive ZLE.
    compadd() {
        [ "$1" = -- ]
        shift
        suggestions=("$@")
    }
else
    [[ "$(complete -p worktree)" == *"-F _worktree_complete_bash worktree" ]]
fi

complete_branches() {
    suggestions=()
    if [ -n "${ZSH_VERSION-}" ]; then
        words=(worktree switch "$1")
        CURRENT=3
        PREFIX="$1"
        _worktree_complete_zsh
    else
        COMP_WORDS=(worktree switch "$1")
        COMP_CWORD=2
        _worktree_complete_bash
        suggestions=("${COMPREPLY[@]}")
    fi
}

complete_branches feature/
[ "${#suggestions[@]}" -eq 2 ]
[ "${suggestions[*]}" = 'feature/one feature/two' ]
complete_branches ''
[ "${suggestions[*]}" = 'feature/one feature/two main' ]
complete_branches missing
[ "${#suggestions[@]}" -eq 0 ]

# Completion also works inside a linked worktree.
worktree switch feature/one
complete_branches feature/t
[ "${suggestions[*]}" = feature/two ]

# Failures stay silent and must not fall back to filenames.
cd "$scratch_directory"
complete_branches '' > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"
[ "${#suggestions[@]}" -eq 0 ]
[ ! -s "$scratch_directory/stdout" ]
[ ! -s "$scratch_directory/stderr" ]
printf 'completion smoke test passed\n'
