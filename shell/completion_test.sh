# Run with both Bash and Zsh from the repository root.
set -e

scratch_directory=$(mktemp -d /tmp/worktree-completion.XXXXXX)
trap 'rm -rf -- "$scratch_directory"' EXIT
go build -o "$scratch_directory/bin/worktree" ./cmd/worktree
export PATH="$scratch_directory/bin:$PATH"

if [ -n "${ZSH_VERSION-}" ]; then
    # Directory hooks must not replace the test binary through PATH changes.
    chpwd_functions=()
    autoload -Uz compinit
    compinit -u -D
    eval "$(worktree init zsh)"
else
    eval "$(worktree init bash)"
fi

git init -q --initial-branch=main "$scratch_directory/repository"
cd "$scratch_directory/repository"
git -c user.name=Test -c user.email=test@example.com -c commit.gpgsign=false commit -q --allow-empty -m initial
git worktree add -q -b feature/one "$scratch_directory/one tree"
git worktree add -q -b feature/two "$scratch_directory/two tree"
git worktree add -q --detach "$scratch_directory/detached"
git branch unused
git remote add origin "$scratch_directory/repository"
git update-ref refs/remotes/origin/remote-only HEAD
git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/remote-only

if [ -n "${ZSH_VERSION-}" ]; then
    [ "${_comps[worktree]}" = _worktree_complete_zsh ]
    # Capture candidates at the completion boundary without an interactive ZLE.
    # This models zsh's real behaviour: by default compadd keeps only candidates
    # that start with the typed word, and -U keeps every candidate. Fuzzy matches
    # reach the shell only when the integration passes -U, so the stub must too.
    compadd() {
        local keep_all=
        while [ "$1" != -- ]; do
            [ "$1" = -U ] && keep_all=1
            shift
        done
        shift
        suggestions=()
        for candidate in "$@"; do
            if [ -n "$keep_all" ] || [ -z "$PREFIX" ] || [ "${candidate#$PREFIX}" != "$candidate" ]; then
                suggestions+=("$candidate")
            fi
        done
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
[ "${suggestions[*]}" = 'feature/one feature/two main origin/remote-only remote-only unused' ]
complete_branches missing
[ "${#suggestions[@]}" -eq 0 ]
# A query matches branches that merely contain it, not only prefixes.
complete_branches sed
[ "${suggestions[*]}" = unused ]

# The standalone command prints the destination without changing the directory.
[ "$(command worktree switch feature/one)" = "$(cd "$scratch_directory/one tree" && pwd -P)" ]

# Completion also works inside a linked worktree.
worktree switch feature/one
complete_branches feature/t
[ "${suggestions[*]}" = feature/two ]

# From a linked worktree, switching to a branch with no worktree creates a
# worktree for it and moves there; the other worktrees keep their branches.
main_checkout="$(cd "$scratch_directory/repository" && pwd -P)"
worktree switch unused
[ "$PWD" = "$(cd "$scratch_directory/repository.worktrees/unused/repository" && pwd -P)" ]
[ "$(git branch --show-current)" = unused ]
[ "$(git -C "$scratch_directory/one tree" branch --show-current)" = feature/one ]
[ "$(git -C "$main_checkout" branch --show-current)" = main ]

# A remote-only name creates a worktree that tracks the remote branch.
worktree switch remote-only
[ "$PWD" = "$(cd "$scratch_directory/repository.worktrees/remote-only/repository" && pwd -P)" ]
[ "$(git branch --show-current)" = remote-only ]
[ "$(git rev-parse --abbrev-ref '@{upstream}')" = origin/remote-only ]

# Invalid arguments keep the current directory.
previous_directory="$PWD"
if worktree switch > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"; then
    exit 1
fi
[ "$(cat "$scratch_directory/stderr")" = 'usage: worktree <switch|list> <branch>' ]
[ "$PWD" = "$previous_directory" ]

# A name with no branch creates a new branch from the current directory.
worktree s added-branch
[ "$(pwd -P)" = "$(cd "$scratch_directory/repository.worktrees/added-branch/repository" && pwd -P)" ]
[ "$(git branch --show-current)" = added-branch ]

# Failures stay silent and must not fall back to filenames.
cd "$scratch_directory"
complete_branches '' > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"
[ "${#suggestions[@]}" -eq 0 ]
[ ! -s "$scratch_directory/stdout" ]
[ ! -s "$scratch_directory/stderr" ]
printf 'completion smoke test passed\n'
