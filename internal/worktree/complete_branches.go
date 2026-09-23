package worktree

import (
	"context"
	"sort"
)

type CompleteCommand struct{ Prefix string }
type CompleteResult struct{ Branches []string }

type BranchLister interface {
	WorktreeLister
	ListBranches(ctx context.Context) ([]string, error)
}

type CompleteBranches struct {
	lister BranchLister
}

func NewCompleteBranches(lister BranchLister) CompleteBranches {
	return CompleteBranches{lister: lister}
}

func (u CompleteBranches) Complete(ctx context.Context, command CompleteCommand) (CompleteResult, error) {
	worktrees, err := u.lister.List(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	candidates, err := u.lister.ListBranches(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	for _, candidate := range worktrees {
		candidates = append(candidates, candidate.Branch())
	}
	seen := make(map[string]bool)
	var branches []string
	scores := make(map[string]int)
	for _, branch := range candidates {
		if branch == "" || seen[branch] {
			continue
		}
		score := fuzzyScore(command.Prefix, branch)
		if command.Prefix != "" && score <= 0 {
			continue
		}
		seen[branch] = true
		scores[branch] = score
		branches = append(branches, branch)
	}
	// Best fuzzy match first; ties, and every branch when no query is typed,
	// fall back to alphabetical order.
	sort.Slice(branches, func(i, j int) bool {
		if scores[branches[i]] != scores[branches[j]] {
			return scores[branches[i]] > scores[branches[j]]
		}
		return branches[i] < branches[j]
	})
	return CompleteResult{Branches: branches}, nil
}
