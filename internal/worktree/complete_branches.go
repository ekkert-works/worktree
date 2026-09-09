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

type completeBranches struct{ lister WorktreeLister }

func NewCompleteBranches(lister WorktreeLister) CompleteBranches {
	return completeBranches{lister: lister}
}

func (u completeBranches) Complete(ctx context.Context, command CompleteCommand) (CompleteResult, error) {
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	seen := make(map[string]bool)
	var branches []string
	for _, candidate := range worktrees {
		branch := candidate.Branch()
		if branch == "" || seen[branch] || !strings.HasPrefix(branch, command.Prefix) {
			continue
		}
		seen[branch] = true
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return CompleteResult{Branches: branches}, nil
}
