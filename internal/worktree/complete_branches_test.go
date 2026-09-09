package worktree_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeBranches struct {
	branches []string
	err      error
}

func (f *fakeBranches) ListBranches(context.Context) ([]string, error) {
	return f.branches, f.err
}

func TestCompleteIncludesBranchesWithoutWorktrees(t *testing.T) {
	main, err := worktree.NewWorktree("/repository", "main")
	if err != nil {
		t.Fatal(err)
	}
	lister := &fakeLister{worktrees: []worktree.Worktree{main}}
	branches := &fakeBranches{branches: []string{"unused", "origin/remote", "remote", "main"}}
	useCase := worktree.NewCompleteBranches(lister, branches)
	result, err := useCase.Complete(context.Background(), worktree.CompleteCommand{})
	want := []string{"main", "origin/remote", "remote", "unused"}
	if err != nil || !reflect.DeepEqual(result.Branches, want) {
		t.Fatalf("got %+v, %v; want %v", result, err, want)
	}
	branches.err = worktree.ErrReadFailure
	_, err = useCase.Complete(context.Background(), worktree.CompleteCommand{})
	if !errors.Is(err, worktree.ErrReadFailure) {
		t.Fatalf("got %v", err)
	}
}

func TestCompleteBranches(t *testing.T) {
	var candidates []worktree.Worktree
	for _, branch := range []string{"main", "feature/two", "", "feature/one", "main"} {
		candidate, err := worktree.NewWorktree("/repository/"+branch, branch)
		if err != nil {
			t.Fatal(err)
		}
		candidates = append(candidates, candidate)
	}
	for _, test := range []struct {
		name   string
		prefix string
		want   []string
	}{
		{name: "all sorted unique branches", want: []string{"feature/one", "feature/two", "main"}},
		{name: "prefix", prefix: "feature/", want: []string{"feature/one", "feature/two"}},
		{name: "exact", prefix: "main", want: []string{"main"}},
		{name: "no matches", prefix: "missing"},
		{name: "literal prefix", prefix: "feature/*"},
	} {
		t.Run(test.name, func(t *testing.T) {
			lister := &fakeLister{worktrees: candidates}
			result, err := worktree.NewCompleteBranches(lister, &fakeBranches{}).Complete(context.Background(), worktree.CompleteCommand{Prefix: test.prefix})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Branches, test.want) {
				t.Fatalf("got %v, want %v", result.Branches, test.want)
			}
		})
	}
}

func TestCompleteReturnsReadFailure(t *testing.T) {
	lister := &fakeLister{err: worktree.ErrReadFailure}
	result, err := worktree.NewCompleteBranches(lister, &fakeBranches{}).Complete(context.Background(), worktree.CompleteCommand{})
	if !errors.Is(err, worktree.ErrReadFailure) || len(result.Branches) != 0 {
		t.Fatalf("got %+v, %v", result, err)
	}
}
