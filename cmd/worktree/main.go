package main

import (
	"context"
	"os"

	"github.com/ekkert-works/worktree/internal/worktree"
	"github.com/ekkert-works/worktree/internal/worktree/cliin"
	"github.com/ekkert-works/worktree/internal/worktree/gitcli"
)

func main() {
	lister := gitcli.NewLister(".")
	checkout := gitcli.NewCheckout(".")
	adder := gitcli.NewAdd(".")
	useCase := worktree.NewSwitchWorktree(lister, checkout)
	addUseCase := worktree.NewAddWorktree(adder, adder)
	completion := worktree.NewCompleteBranches(lister)
	os.Exit(cliin.Run(context.Background(), os.Args[1:], useCase, addUseCase, completion, lister, os.Stdout, os.Stderr))
}
