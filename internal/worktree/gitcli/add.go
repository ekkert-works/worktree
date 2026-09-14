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

func (a *Add) Add(ctx context.Context, path, branch string) error {
	_, err := runGit(ctx, a.directory, worktree.ErrAddFailure, "worktree", "add", path, branch)
	return err
}
