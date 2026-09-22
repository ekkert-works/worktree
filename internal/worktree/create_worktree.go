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

type CreateCommand struct{ Branch string }
type CreateResult struct{ Path string }

type CreateWorktree struct {
	main  MainCheckoutReader
	adder WorktreeAdder
}

func NewCreateWorktree(main MainCheckoutReader, adder WorktreeAdder) CreateWorktree {
	return CreateWorktree{main: main, adder: adder}
}

// Create adds a worktree under <main>.worktrees/<branch>/<repository> and
// returns its path. The repository name is the leaf so tty environments that
// show the current directory keep naming the repository, not the branch. The
// adapter checks out an existing branch or creates a new one. Git rejects a
// branch that is already checked out in another worktree.
func (u CreateWorktree) Create(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if command.Branch == "" || strings.HasPrefix(command.Branch, "-") {
		return CreateResult{}, ErrInvalidBranch
	}
	root, err := u.main.ReadMainCheckout(ctx)
	if err != nil {
		return CreateResult{}, err
	}
	repository := root[strings.LastIndex(root, "/")+1:]
	destination := root + ".worktrees/" + command.Branch + "/" + repository
	if err := u.adder.Add(ctx, destination, command.Branch); err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Path: destination}, nil
}
