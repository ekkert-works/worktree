package worktree

import (
	"context"
	"sort"
	"strings"
)

type CompleteBranches interface {
	Complete(ctx context.Context, command CompleteCommand) (CompleteResult, error)
}

type CompleteCommand struct{ Prefix string }
type CompleteResult struct{ Branches []string }

type BranchLister interface {
	ListBranches(ctx context.Context) ([]string, error)
}

type completeBranches struct {
	lister   WorktreeLister
	branches BranchLister
}

func NewCompleteBranches(lister WorktreeLister, branches BranchLister) CompleteBranches {
	return completeBranches{lister: lister, branches: branches}
}

func (u completeBranches) Complete(ctx context.Context, command CompleteCommand) (CompleteResult, error) {
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	candidates, err := u.branches.ListBranches(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	for _, candidate := range worktrees {
		candidates = append(candidates, candidate.Branch())
	}
	seen := make(map[string]bool)
	var branches []string
	for _, branch := range candidates {
		if branch == "" || seen[branch] || !strings.HasPrefix(branch, command.Prefix) {
			continue
		}
		seen[branch] = true
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return CompleteResult{Branches: branches}, nil
}
