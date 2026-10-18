package worktree

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidBranch = errors.New("worktree: invalid branch name")
	ErrAmbiguous     = errors.New("worktree: multiple worktrees for branch")
	ErrUnavailable   = errors.New("worktree: git is unavailable")
	ErrReadFailure   = errors.New("worktree: cannot read worktrees")
)

type WorktreeLister interface {
	List(ctx context.Context) ([]Worktree, error)
}

// WorktreeCreator adds a worktree for the branch and returns its destination.
// An existing branch is checked out; an unknown name creates the branch from
// the one checked out in the current directory.
type WorktreeCreator interface {
	Create(ctx context.Context, command CreateCommand) (CreateResult, error)
}

type SwitchCommand struct{ Branch string }
type SwitchResult struct{ Path string }

type SwitchWorktree struct {
	lister  WorktreeLister
	creator WorktreeCreator
}

func NewSwitchWorktree(lister WorktreeLister, creator WorktreeCreator) SwitchWorktree {
	return SwitchWorktree{lister: lister, creator: creator}
}

// Switch moves to the worktree that holds the branch. The branch name must
// match exactly; fuzzy matching is left to shell completion, where the user
// sees and picks the candidate. When no worktree matches, it
// creates one: an existing branch is checked out, and an unknown name creates a
// branch from the one checked out in the current directory. Worktrees are the
// only unit this tool works with; it never checks a branch out in place.
func (u SwitchWorktree) Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error) {
	if command.Branch == "" || strings.HasPrefix(command.Branch, "-") {
		return SwitchResult{}, ErrInvalidBranch
	}
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return SwitchResult{}, err
	}
	path, matched, err := resolveWorktree(worktrees, command.Branch)
	if err != nil {
		return SwitchResult{}, err
	}
	if matched {
		return SwitchResult{Path: path}, nil
	}
	created, err := u.creator.Create(ctx, CreateCommand{Branch: command.Branch})
	if err != nil {
		return SwitchResult{}, err
	}
	return SwitchResult{Path: created.Path}, nil
}

// resolveWorktree finds the worktree that holds the branch. It reports
// matched=false when no worktree does, so the caller can fall back to creating
// one. More than one worktree for the branch is ambiguous.
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
	return path, matched, nil
}
