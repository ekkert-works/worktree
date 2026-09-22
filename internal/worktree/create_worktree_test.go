package worktree_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeAdd struct {
	root      string
	rootError error
	addError  error
	path      string
	branch    string
	addCalls  int
}

func (f *fakeAdd) ReadMainCheckout(context.Context) (string, error) {
	return f.root, f.rootError
}

func (f *fakeAdd) Add(_ context.Context, path, branch string) error {
	f.addCalls++
	f.path = path
	f.branch = branch
	return f.addError
}

func TestCreate(t *testing.T) {
	for _, test := range []struct {
		name      string
		branch    string
		rootError error
		addError  error
		wantError error
		wantPath  string
		wantCalls int
	}{
		{name: "created", branch: "feature/x", wantPath: "/main.worktrees/feature/x/main", wantCalls: 1},
		{name: "empty", wantError: worktree.ErrInvalidBranch},
		{name: "option", branch: "--force", wantError: worktree.ErrInvalidBranch},
		{name: "root failure", branch: "feature", rootError: worktree.ErrReadFailure, wantError: worktree.ErrReadFailure},
		{name: "add failure", branch: "feature", addError: worktree.ErrAddFailure, wantError: worktree.ErrAddFailure, wantCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			adder := &fakeAdd{root: "/main", rootError: test.rootError, addError: test.addError}
			result, err := worktree.NewCreateWorktree(adder, adder).Create(context.Background(), worktree.CreateCommand{Branch: test.branch})
			if !errors.Is(err, test.wantError) || result.Path != test.wantPath || adder.addCalls != test.wantCalls {
				t.Fatalf("got %+v, %v, %d add calls", result, err, adder.addCalls)
			}
			if err == nil && adder.branch != test.branch {
				t.Fatalf("added branch %q, want %q", adder.branch, test.branch)
			}
		})
	}
}
