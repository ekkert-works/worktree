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
	os.Exit(cliin.RunGitWT(context.Background(), os.Args[1:], useCase, os.Stdout, os.Stderr))
}
