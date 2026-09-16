package gitcli

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type Add struct{ directory string }

func NewAdd(directory string) *Add {
	return &Add{directory: directory}
}

// ReadMainCheckout returns the main worktree path, the parent of the shared
// git directory. This holds for standard repositories, not bare ones.
func (a *Add) ReadMainCheckout(ctx context.Context) (string, error) {
	commonDirectory, err := runGit(ctx, a.directory, worktree.ErrReadFailure,
		"rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSuffix(commonDirectory, "\n")), nil
}

// Add checks out an existing branch, or creates the branch from HEAD when no
// local ref matches. A name that exists only as a remote-tracking branch is
// treated as new; it does not set up tracking.
func (a *Add) Add(ctx context.Context, path, branch string) error {
	arguments := []string{"worktree", "add", path, branch}
	if !a.branchExists(ctx, branch) {
		arguments = []string{"worktree", "add", "-b", branch, path}
	}
	_, err := runGit(ctx, a.directory, worktree.ErrAddFailure, arguments...)
	return err
}

func (a *Add) branchExists(ctx context.Context, branch string) bool {
	_, err := runGit(ctx, a.directory, worktree.ErrAddFailure,
		"show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}
