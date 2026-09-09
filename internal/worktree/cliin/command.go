package cliin

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ekkert-works/worktree/internal/worktree"
)

// Run writes only the destination to stdout. Errors and usage go to stderr.
func Run(ctx context.Context, arguments []string, useCase worktree.SwitchWorktree, stdout, stderr io.Writer) int {
	if len(arguments) != 2 || arguments[0] != "switch" {
		fmt.Fprintln(stderr, "usage: worktree switch <branch>")
		return 2
	}
	result, err := useCase.Switch(ctx, worktree.SwitchCommand{Branch: arguments[1]})
	switch {
	case errors.Is(err, worktree.ErrInvalidBranch):
		fmt.Fprintln(stderr, err)
		return 2
	case err != nil:
		fmt.Fprintln(stderr, err)
		return 1
	}
	if _, err := fmt.Fprintln(stdout, result.Path); err != nil {
		fmt.Fprintln(stderr, "worktree: cannot write destination")
		return 1
	}
	return 0
}
