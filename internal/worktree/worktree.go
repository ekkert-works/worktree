package worktree

type Worktree struct {
	path   string
	branch string
}

// NewWorktree also accepts detached worktrees, which have no branch.
func NewWorktree(path, branch string) Worktree {
	return Worktree{path: path, branch: branch}
}

func (w Worktree) Path() string { return w.path }

func (w Worktree) Branch() string { return w.branch }

func (w Worktree) MatchesBranch(branch string) bool {
	return w.branch == branch
}
