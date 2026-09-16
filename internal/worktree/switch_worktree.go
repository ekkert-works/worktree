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
	ErrCheckoutFailure = errors.New("worktree: cannot check out branch; check the branch name and local changes")
)

type WorktreeLister interface {
	List(ctx context.Context) ([]Worktree, error)
}

type Checkout interface {
	// MainCheckout reports the absolute path of the repository's main checkout.
	MainCheckout(ctx context.Context) (string, error)
	Checkout(ctx context.Context, directory, branch string) error
}

type SwitchCommand struct{ Branch string }
type SwitchResult struct{ Path string }

type SwitchWorktree struct {
	lister   WorktreeLister
	checkout Checkout
}

func NewSwitchWorktree(lister WorktreeLister, checkout Checkout) SwitchWorktree {
	return SwitchWorktree{lister: lister, checkout: checkout}
}

func (u SwitchWorktree) Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error) {
	if command.Branch == "" || strings.HasPrefix(command.Branch, "-") {
		return SwitchResult{}, ErrInvalidBranch
	}
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return SwitchResult{}, err
	}
	if path, matched, err := resolveWorktree(worktrees, command.Branch); err != nil {
		return SwitchResult{}, err
	} else if matched {
		return SwitchResult{Path: path}, nil
	}
	// No worktree holds the branch, so check it out in the main checkout. From a
	// linked worktree this moves to the main checkout instead of changing the
	// current worktree's branch.
	mainCheckout, err := u.checkout.MainCheckout(ctx)
	if err != nil {
		return SwitchResult{}, err
	}
	if err := u.checkout.Checkout(ctx, mainCheckout, command.Branch); err != nil {
		return SwitchResult{}, err
	}
	return SwitchResult{Path: mainCheckout}, nil
}

// resolveWorktree finds the worktree to switch to. An exact branch match wins;
// otherwise the highest-scoring fuzzy match does. It reports matched=false when
// no worktree matches, so the caller can fall back to a literal checkout. A tie
// for the best match, exact or fuzzy, is ambiguous.
func resolveWorktree(worktrees []Worktree, query string) (path string, matched bool, err error) {
	for _, candidate := range worktrees {
		if !candidate.MatchesBranch(query) {
			continue
		}
		if matched {
			return "", false, ErrAmbiguous
		}
		path, matched = candidate.Path(), true
	}
	if matched {
		return path, true, nil
	}

	bestScore := 0
	tied := false
	for _, candidate := range worktrees {
		score := fuzzyScore(query, candidate.Branch())
		switch {
		case score <= 0:
		case score > bestScore:
			path, bestScore, matched, tied = candidate.Path(), score, true, false
		case score == bestScore:
			tied = true
		}
	}
	if tied {
		return "", false, ErrAmbiguous
	}
	return path, matched, nil
}
