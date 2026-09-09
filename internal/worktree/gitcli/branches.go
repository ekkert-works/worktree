package gitcli

import (
	"context"
	"strings"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func (l *Lister) ListBranches(ctx context.Context) ([]string, error) {
	output, err := runGit(ctx, l.directory, worktree.ErrReadFailure,
		"for-each-ref", "--format=%(refname)%09%(symref)", "refs/heads/", "refs/remotes/")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(output, "\n") {
		reference, symbolic, found := strings.Cut(line, "\t")
		if !found || symbolic != "" {
			continue
		}
		if branch, local := strings.CutPrefix(reference, "refs/heads/"); local {
			branches = append(branches, branch)
			continue
		}
		remoteBranch := strings.TrimPrefix(reference, "refs/remotes/")
		branches = append(branches, remoteBranch)
		// Git checkout can create a local branch from a matching remote branch.
		_, branch, found := strings.Cut(remoteBranch, "/")
		if found {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}
