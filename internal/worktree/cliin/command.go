package cliin

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type Switcher interface {
	Switch(ctx context.Context, command worktree.SwitchCommand) (worktree.SwitchResult, error)
}

type Adder interface {
	Add(ctx context.Context, command worktree.AddCommand) (worktree.AddResult, error)
}

type Completer interface {
	Complete(ctx context.Context, command worktree.CompleteCommand) (worktree.CompleteResult, error)
}

type Lister interface {
	List(ctx context.Context) ([]worktree.Worktree, error)
}

// Run writes destinations or completion candidates to stdout.
func Run(ctx context.Context, arguments []string, useCase Switcher, adder Adder, completion Completer, lister Lister, stdout, stderr io.Writer) int {
	if len(arguments) > 0 && arguments[0] == "__complete" {
		return complete(ctx, arguments[1:], completion, stdout)
	}
	if len(arguments) > 0 && arguments[0] == "init" {
		return initShell(arguments[1:], stdout, stderr)
	}
	if len(arguments) == 1 && arguments[0] == "list" {
		return listWorktrees(ctx, lister, stdout, stderr)
	}
	if len(arguments) == 2 && arguments[0] == "add" {
		return addBranch(ctx, arguments[1], adder, stdout, stderr)
	}
	if len(arguments) != 2 || !isSwitch(arguments[0]) {
		fmt.Fprintln(stderr, "usage: worktree <switch|add|list> <branch>")
		return 2
	}
	return switchBranch(ctx, arguments[1], useCase, stdout, stderr)
}

func listWorktrees(ctx context.Context, useCase Lister, stdout, stderr io.Writer) int {
	worktrees, err := useCase.List(ctx)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, tree := range worktrees {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\n", tree.Branch(), tree.Path()); err != nil {
			fmt.Fprintln(stderr, "worktree: cannot write worktree list")
			return 1
		}
	}
	return 0
}

func addBranch(ctx context.Context, branch string, useCase Adder, stdout, stderr io.Writer) int {
	result, err := useCase.Add(ctx, worktree.AddCommand{Branch: branch})
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

func isSwitch(command string) bool {
	return command == "switch" || command == "s"
}

func switchBranch(ctx context.Context, branch string, useCase Switcher, stdout, stderr io.Writer) int {
	result, err := useCase.Switch(ctx, worktree.SwitchCommand{Branch: branch})
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
func complete(ctx context.Context, arguments []string, useCase Completer, stdout io.Writer) int {
	if len(arguments) != 2 || !isSwitch(arguments[0]) {
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
