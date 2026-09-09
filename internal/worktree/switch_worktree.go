package worktree

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidBranch   = errors.New("worktree: invalid branch name")
	ErrAmbiguous       = errors.New("worktree: multiple worktrees for branch")
	ErrUnavailable     = errors.New("worktree: git is unavailable")
	ErrReadFailure     = errors.New("worktree: cannot read worktrees")
	ErrLinkedCheckout  = errors.New("worktree: cannot check out a branch in a linked worktree; switch from the main checkout")
	ErrCheckoutFailure = errors.New("worktree: cannot check out branch; check the branch name and local changes")
)

type WorktreeLister interface {
	List(ctx context.Context) ([]Worktree, error)
}

type CheckoutLocation struct {
	Path   string
	Linked bool
}

type CheckoutLocationReader interface {
	ReadLocation(ctx context.Context) (CheckoutLocation, error)
}

type BranchCheckout interface {
	Checkout(ctx context.Context, branch string) error
}

type SwitchWorktree interface {
	Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error)
}

type SwitchCommand struct{ Branch string }
type SwitchResult struct{ Path string }

type switchWorktree struct {
	lister   WorktreeLister
	location CheckoutLocationReader
	checkout BranchCheckout
}

func NewSwitchWorktree(lister WorktreeLister, location CheckoutLocationReader, checkout BranchCheckout) SwitchWorktree {
	return switchWorktree{lister: lister, location: location, checkout: checkout}
}

func (u switchWorktree) Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error) {
	if command.Branch == "" || strings.HasPrefix(command.Branch, "-") {
		return SwitchResult{}, ErrInvalidBranch
	}
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return SwitchResult{}, err
	}
	var result SwitchResult
	for _, candidate := range worktrees {
		if !candidate.MatchesBranch(command.Branch) {
			continue
		}
		if result.Path != "" {
			return SwitchResult{}, ErrAmbiguous
		}
		result.Path = candidate.Path()
	}
	if result.Path != "" {
		return result, nil
	}
	location, err := u.location.ReadLocation(ctx)
	if err != nil {
		return SwitchResult{}, err
	}
	if location.Linked {
		return SwitchResult{}, ErrLinkedCheckout
	}
	if err := u.checkout.Checkout(ctx, command.Branch); err != nil {
		return SwitchResult{}, err
	}
	return SwitchResult{Path: location.Path}, nil
}
