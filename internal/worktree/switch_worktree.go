package worktree

import (
	"context"
	"errors"
)

var (
	ErrInvalidBranch = errors.New("worktree: branch is required")
	ErrNotFound      = errors.New("worktree: no worktree for branch")
	ErrAmbiguous     = errors.New("worktree: multiple worktrees for branch")
	ErrUnavailable   = errors.New("worktree: git is unavailable")
	ErrReadFailure   = errors.New("worktree: cannot read worktrees")
)

type WorktreeLister interface {
	List(ctx context.Context) ([]Worktree, error)
}

type SwitchWorktree interface {
	Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error)
}

type SwitchCommand struct{ Branch string }
type SwitchResult struct{ Path string }

type switchWorktree struct{ lister WorktreeLister }

func NewSwitchWorktree(lister WorktreeLister) SwitchWorktree {
	return switchWorktree{lister: lister}
}

func (u switchWorktree) Switch(ctx context.Context, command SwitchCommand) (SwitchResult, error) {
	if command.Branch == "" {
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
	if result.Path == "" {
		return SwitchResult{}, ErrNotFound
	}
	return result, nil
}
