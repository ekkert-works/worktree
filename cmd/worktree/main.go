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
	useCase := worktree.NewSwitchWorktree(lister)
	os.Exit(cliin.Run(context.Background(), os.Args[1:], useCase, os.Stdout, os.Stderr))
}
