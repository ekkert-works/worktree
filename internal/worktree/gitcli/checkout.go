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

func (c *Checkout) ReadLocation(ctx context.Context) (worktree.CheckoutLocation, error) {
	gitDirectory, err := runGit(ctx, c.directory, worktree.ErrReadFailure,
		"rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return worktree.CheckoutLocation{}, err
	}
	commonDirectory, err := runGit(ctx, c.directory, worktree.ErrReadFailure,
		"rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return worktree.CheckoutLocation{}, err
	}
	path, err := filepath.Abs(c.directory)
	if err != nil {
		return worktree.CheckoutLocation{}, worktree.ErrReadFailure
	}
	return worktree.CheckoutLocation{
		Path:   path,
		Linked: strings.TrimSuffix(gitDirectory, "\n") != strings.TrimSuffix(commonDirectory, "\n"),
	}, nil
}

func (c *Checkout) Checkout(ctx context.Context, branch string) error {
	// The separator prevents Git from treating a missing branch as a file path.
	_, err := runGit(ctx, c.directory, worktree.ErrCheckoutFailure, "checkout", branch, "--")
	return err
}
