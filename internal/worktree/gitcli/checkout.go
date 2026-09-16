package gitcli

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type Checkout struct{ directory string }

func NewCheckout(directory string) *Checkout {
	return &Checkout{directory: directory}
}

func (c *Checkout) MainCheckout(ctx context.Context) (string, error) {
	// The common Git directory is the main checkout's ".git"; its parent is the
	// main checkout, whether the current directory is the main or a linked worktree.
	commonDirectory, err := runGit(ctx, c.directory, worktree.ErrReadFailure,
		"rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSuffix(commonDirectory, "\n")), nil
}

func (c *Checkout) Checkout(ctx context.Context, directory, branch string) error {
	// The separator prevents Git from treating a missing branch as a file path.
	_, err := runGit(ctx, directory, worktree.ErrCheckoutFailure, "checkout", branch, "--")
	return err
}
