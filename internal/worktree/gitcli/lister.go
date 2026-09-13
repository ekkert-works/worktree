package gitcli

import (
	"context"
	"strings"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type Lister struct{ directory string }

func NewLister(directory string) *Lister {
	return &Lister{directory: directory}
}

func (l *Lister) List(ctx context.Context) ([]worktree.Worktree, error) {
	output, err := runGit(ctx, l.directory, worktree.ErrReadFailure,
		"worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	return parse(output)
}

func parse(output string) ([]worktree.Worktree, error) {
	var result []worktree.Worktree
	for _, record := range strings.Split(output, "\x00\x00") {
		if record == "" {
			continue
		}
		var path, branch string
		var bare bool
		for _, field := range strings.Split(record, "\x00") {
			switch {
			case strings.HasPrefix(field, "worktree "):
				path = strings.TrimPrefix(field, "worktree ")
			case strings.HasPrefix(field, "branch refs/heads/"):
				branch = strings.TrimPrefix(field, "branch refs/heads/")
			case field == "bare":
				bare = true
			}
		}
		if bare {
			continue
		}
		candidate, err := worktree.NewWorktree(path, branch)
		if err != nil {
			return nil, worktree.ErrReadFailure
		}
		result = append(result, candidate)
	}
	return result, nil
}
