package worktree_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeLister struct {
	worktrees []worktree.Worktree
	err       error
	calls     int
}

func (f *fakeLister) List(context.Context) ([]worktree.Worktree, error) {
	f.calls++
	return f.worktrees, f.err
}

func TestSwitch(t *testing.T) {
	feature, err := worktree.NewWorktree("/repo with spaces", "feature")
	if err != nil {
		t.Fatal(err)
	}
	detached, err := worktree.NewWorktree("/detached", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name      string
		branch    string
		worktrees []worktree.Worktree
		listError error
		wantError error
		wantPath  string
	}{
		{name: "match", branch: "feature", worktrees: []worktree.Worktree{detached, feature}, wantPath: "/repo with spaces"},
		{name: "empty", wantError: worktree.ErrInvalidBranch},
		{name: "option", branch: "--force", wantError: worktree.ErrInvalidBranch},
		{name: "missing", branch: "other", worktrees: []worktree.Worktree{feature}, wantError: worktree.ErrCheckoutFailure},
		{name: "exact match only", branch: "feat", worktrees: []worktree.Worktree{feature}, wantError: worktree.ErrCheckoutFailure},
		{name: "ambiguous", branch: "feature", worktrees: []worktree.Worktree{feature, feature}, wantError: worktree.ErrAmbiguous},
		{name: "read failure", branch: "feature", listError: worktree.ErrReadFailure, wantError: worktree.ErrReadFailure},
	} {
		t.Run(test.name, func(t *testing.T) {
			lister := &fakeLister{worktrees: test.worktrees, err: test.listError}
			checkout := &fakeCheckout{checkoutError: worktree.ErrCheckoutFailure}
			result, err := worktree.NewSwitchWorktree(lister, checkout, checkout).Switch(context.Background(), worktree.SwitchCommand{Branch: test.branch})
			if !errors.Is(err, test.wantError) || result.Path != test.wantPath {
				t.Fatalf("got %+v, %v; want path %q, %v", result, err, test.wantPath, test.wantError)
			}
			if test.wantError != worktree.ErrCheckoutFailure && (checkout.locationCalls != 0 || checkout.checkoutCalls != 0) {
				t.Fatal("checkout reached before resolving existing worktrees")
			}
			if test.wantError == worktree.ErrInvalidBranch && lister.calls != 0 {
				t.Fatal("invalid input reached lister")
			}
		})
	}
}

func TestRejectEmptyPath(t *testing.T) {
	_, err := worktree.NewWorktree("", "feature")
	if !errors.Is(err, worktree.ErrInvalidWorktree) {
		t.Fatalf("got %v", err)
	}
}
