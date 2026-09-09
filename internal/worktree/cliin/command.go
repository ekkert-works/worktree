package cliin

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ekkert-works/worktree/internal/worktree"
)

// Run writes destinations or completion candidates to stdout.
func Run(ctx context.Context, arguments []string, useCase worktree.SwitchWorktree, completion worktree.CompleteBranches, stdout, stderr io.Writer) int {
	if len(arguments) > 0 && arguments[0] == "__complete" {
		return complete(ctx, arguments[1:], completion, stdout)
	}
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

// Completion is a private shell protocol: one branch per line, no diagnostics.
func complete(ctx context.Context, arguments []string, useCase worktree.CompleteBranches, stdout io.Writer) int {
	if len(arguments) != 2 || arguments[0] != "switch" {
		return 2
	}
	result, err := useCase.Complete(ctx, worktree.CompleteCommand{Prefix: arguments[1]})
	if err != nil {
		return 1
	}
	for _, branch := range result.Branches {
		if _, err := fmt.Fprintln(stdout, branch); err != nil {
			return 1
		}
	}
	return 0
}
