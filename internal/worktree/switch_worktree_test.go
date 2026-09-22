package worktree_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeLister struct {
	worktrees []worktree.Worktree
	branches  []string
	err       error
	branchErr error
	calls     int
}

func (f *fakeLister) List(context.Context) ([]worktree.Worktree, error) {
	f.calls++
	return f.worktrees, f.err
}

func (f *fakeLister) ListBranches(context.Context) ([]string, error) {
	return f.branches, f.branchErr
}

// fakeCreator stands in for the add-worktree use case the switch falls back to.
type fakeCreator struct {
	path   string
	err    error
	branch string
	calls  int
}

func (f *fakeCreator) Create(_ context.Context, command worktree.CreateCommand) (worktree.CreateResult, error) {
	f.calls++
	f.branch = command.Branch
	return worktree.CreateResult{Path: f.path}, f.err
}

func TestSwitch(t *testing.T) {
	feature := worktree.NewWorktree("/repo with spaces", "feature")
	detached := worktree.NewWorktree("/detached", "")
	for _, test := range []struct {
		name        string
		branch      string
		worktrees   []worktree.Worktree
		listError   error
		createError error
		wantError   error
		wantPath    string
		wantCreate  bool
	}{
		{name: "match", branch: "feature", worktrees: []worktree.Worktree{detached, feature}, wantPath: "/repo with spaces"},
		{name: "empty", wantError: worktree.ErrInvalidBranch},
		{name: "option", branch: "--force", wantError: worktree.ErrInvalidBranch},
		{name: "no match creates worktree", branch: "xyz", worktrees: []worktree.Worktree{feature}, wantPath: "/created", wantCreate: true},
		{name: "create failure", branch: "xyz", worktrees: []worktree.Worktree{feature}, createError: worktree.ErrAddFailure, wantError: worktree.ErrAddFailure, wantCreate: true},
		{name: "fuzzy match", branch: "feat", worktrees: []worktree.Worktree{feature}, wantPath: "/repo with spaces"},
		{name: "fuzzy ranks best", branch: "feat", worktrees: []worktree.Worktree{worktree.NewWorktree("/other", "some-feature-thing"), feature}, wantPath: "/repo with spaces"},
		{name: "ambiguous", branch: "feature", worktrees: []worktree.Worktree{feature, feature}, wantError: worktree.ErrAmbiguous},
		{name: "fuzzy tie is ambiguous", branch: "ab", worktrees: []worktree.Worktree{worktree.NewWorktree("/x", "a-b-1"), worktree.NewWorktree("/y", "a-b-2")}, wantError: worktree.ErrAmbiguous},
		{name: "read failure", branch: "feature", listError: worktree.ErrReadFailure, wantError: worktree.ErrReadFailure},
	} {
		t.Run(test.name, func(t *testing.T) {
			lister := &fakeLister{worktrees: test.worktrees, err: test.listError}
			creator := &fakeCreator{path: "/created", err: test.createError}
			result, err := worktree.NewSwitchWorktree(lister, creator).Switch(context.Background(), worktree.SwitchCommand{Branch: test.branch})
			if !errors.Is(err, test.wantError) || result.Path != test.wantPath {
				t.Fatalf("got %+v, %v; want path %q, %v", result, err, test.wantPath, test.wantError)
			}
			if (creator.calls != 0) != test.wantCreate {
				t.Fatalf("create called %d times; want create %v", creator.calls, test.wantCreate)
			}
			if test.wantCreate && creator.branch != test.branch {
				t.Fatalf("created branch %q, want %q", creator.branch, test.branch)
			}
			if test.wantError == worktree.ErrInvalidBranch && lister.calls != 0 {
				t.Fatal("invalid input reached lister")
			}
		})
	}
}
