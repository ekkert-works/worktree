# Run with both Bash and Zsh from the repository root.
set -e

project_directory="$PWD"
scratch_directory=$(mktemp -d /tmp/worktree-completion.XXXXXX)
trap 'rm -rf -- "$scratch_directory"' EXIT
go build -o "$scratch_directory/bin/worktree" ./cmd/worktree
go build -o "$scratch_directory/bin/git-wt" ./cmd/git-wt
export PATH="$scratch_directory/bin:$PATH"

if [ -n "${ZSH_VERSION-}" ]; then
    # Directory hooks must not replace the test binary through PATH changes.
    chpwd_functions=()
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
git remote add origin "$scratch_directory/repository"
git update-ref refs/remotes/origin/remote-only HEAD
git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/remote-only

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
[ "${suggestions[*]}" = 'feature/one feature/two main origin/remote-only remote-only unused' ]
complete_branches missing
[ "${#suggestions[@]}" -eq 0 ]

# Both Git completion providers use the same branch list.
if [ -n "${ZSH_VERSION-}" ]; then
    words=(wt feature/)
    CURRENT=2
    PREFIX=feature/
    _git-wt
    [ "${suggestions[*]}" = 'feature/one feature/two' ]
fi
__gitcomp_nl() {
    git_suggestions="$1"
}
cword=2
cur=feature/
_git_wt
[ "$git_suggestions" = "$(printf 'feature/one\nfeature/two')" ]

# Git finds the standalone command through PATH; it prints the destination.
[ "$(command git wt feature/one)" = "$(cd "$scratch_directory/one tree" && pwd -P)" ]

# Completion also works inside a linked worktree.
git wt feature/one
complete_branches feature/t
[ "${suggestions[*]}" = feature/two ]

# A linked worktree must keep its branch when the target has no worktree.
previous_directory="$PWD"
if git wt unused > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"; then
    exit 1
fi
[ "$PWD" = "$previous_directory" ]
[ "$(git branch --show-current)" = feature/one ]
[ ! -s "$scratch_directory/stdout" ]
[[ "$(cat "$scratch_directory/stderr")" == *'linked worktree'* ]]

# In the main checkout, local and remote-only branches use Git checkout.
git wt main
previous_directory="$PWD"
git wt unused
[ "$PWD" = "$previous_directory" ]
[ "$(git branch --show-current)" = unused ]
worktree switch remote-only
[ "$PWD" = "$previous_directory" ]
[ "$(git branch --show-current)" = remote-only ]
[ "$(git rev-parse --abbrev-ref '@{upstream}')" = origin/remote-only ]

# Invalid Git subcommand arguments keep the current directory and branch.
if git wt > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"; then
    exit 1
fi
[ "$(cat "$scratch_directory/stderr")" = 'usage: git wt <branch>' ]
[ "$PWD" = "$previous_directory" ]
[ "$(git branch --show-current)" = remote-only ]

# Failures stay silent and must not fall back to filenames.
cd "$scratch_directory"
complete_branches '' > "$scratch_directory/stdout" 2> "$scratch_directory/stderr"
[ "${#suggestions[@]}" -eq 0 ]
[ ! -s "$scratch_directory/stdout" ]
[ ! -s "$scratch_directory/stderr" ]
printf 'completion smoke test passed\n'
