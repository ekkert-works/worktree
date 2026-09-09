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
	useCase := worktree.NewSwitchWorktree(lister, checkout, checkout)
	completion := worktree.NewCompleteBranches(lister, lister)
	os.Exit(cliin.Run(context.Background(), os.Args[1:], useCase, completion, os.Stdout, os.Stderr))
}
