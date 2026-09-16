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

// Add checks out an existing branch, or creates the branch from HEAD when the
// name matches no ref. A name that exists only as a remote-tracking branch is
// created through Git's own DWIM, which sets up tracking.
func (a *Add) Add(ctx context.Context, path, branch string) error {
	arguments := []string{"worktree", "add", path, branch}
	if !a.knownBranch(ctx, branch) {
		arguments = []string{"worktree", "add", "-b", branch, path}
	}
	_, err := runGit(ctx, a.directory, worktree.ErrAddFailure, arguments...)
	return err
}

// knownBranch reports whether Git can resolve the name to a local branch or a
// remote-tracking branch. In both cases the plain "worktree add <path> <branch>"
// form checks out or creates and tracks it.
func (a *Add) knownBranch(ctx context.Context, branch string) bool {
	if err := a.showRef(ctx, "refs/heads/"+branch); err == nil {
		return true
	}
	output, err := runGit(ctx, a.directory, worktree.ErrAddFailure,
		"for-each-ref", "--format=x", "refs/remotes/*/"+branch)
	return err == nil && strings.TrimSpace(output) != ""
}

func (a *Add) showRef(ctx context.Context, ref string) error {
	_, err := runGit(ctx, a.directory, worktree.ErrAddFailure, "show-ref", "--verify", "--quiet", ref)
	return err
}
