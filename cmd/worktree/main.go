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
	adder := gitcli.NewAdd(".")
	addUseCase := worktree.NewAddWorktree(adder, adder)
	switchUseCase := worktree.NewSwitchWorktree(lister, addUseCase)
	completion := worktree.NewCompleteBranches(lister)
	os.Exit(cliin.Run(context.Background(), os.Args[1:], switchUseCase, completion, lister, os.Stdout, os.Stderr))
}
