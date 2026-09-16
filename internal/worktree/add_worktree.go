package worktree

import (
	"context"
	"errors"
	"strings"
)

var ErrAddFailure = errors.New("worktree: cannot add worktree; check the branch name and whether it is already checked out")

type MainCheckoutReader interface {
	ReadMainCheckout(ctx context.Context) (string, error)
}

type WorktreeAdder interface {
	Add(ctx context.Context, path, branch string) error
}

type AddCommand struct{ Branch string }
type AddResult struct{ Path string }

type AddWorktree struct {
	main  MainCheckoutReader
	adder WorktreeAdder
}

func NewAddWorktree(main MainCheckoutReader, adder WorktreeAdder) AddWorktree {
	return AddWorktree{main: main, adder: adder}
}

// Add creates a worktree under <main>.worktrees/<branch> and returns its path.
// The adapter checks out an existing branch or creates a new one. Git rejects
// a branch that is already checked out in another worktree.
func (u AddWorktree) Add(ctx context.Context, command AddCommand) (AddResult, error) {
	if command.Branch == "" || strings.HasPrefix(command.Branch, "-") {
		return AddResult{}, ErrInvalidBranch
	}
	root, err := u.main.ReadMainCheckout(ctx)
	if err != nil {
		return AddResult{}, err
	}
	destination := root + ".worktrees/" + command.Branch
	if err := u.adder.Add(ctx, destination, command.Branch); err != nil {
		return AddResult{}, err
	}
	return AddResult{Path: destination}, nil
}
