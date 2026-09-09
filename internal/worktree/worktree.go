package worktree

import "errors"

var ErrInvalidWorktree = errors.New("worktree: invalid path")

type Worktree struct {
	path   string
	branch string
}

// NewWorktree also accepts detached worktrees, which have no branch.
func NewWorktree(path, branch string) (Worktree, error) {
	if path == "" {
		return Worktree{}, ErrInvalidWorktree
	}
	return Worktree{path: path, branch: branch}, nil
}

func (w Worktree) Path() string { return w.path }

func (w Worktree) Branch() string { return w.branch }

func (w Worktree) MatchesBranch(branch string) bool {
	return branch != "" && w.branch == branch
}
