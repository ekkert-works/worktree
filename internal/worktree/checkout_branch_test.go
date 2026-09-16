package worktree_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeCheckout struct {
	mainCheckout  string
	mainError     error
	checkoutError error
	mainCalls     int
	checkoutCalls int
	directory     string
	branch        string
}

func (f *fakeCheckout) MainCheckout(context.Context) (string, error) {
	f.mainCalls++
	return f.mainCheckout, f.mainError
}

func (f *fakeCheckout) Checkout(_ context.Context, directory, branch string) error {
	f.checkoutCalls++
	f.directory = directory
	f.branch = branch
	return f.checkoutError
}

func TestCheckoutWithoutWorktree(t *testing.T) {
	for _, test := range []struct {
		name          string
		mainError     error
		checkoutError error
		wantError     error
		wantCalls     int
	}{
		{name: "checks out in main checkout", wantCalls: 1},
		{name: "main lookup failure", mainError: worktree.ErrReadFailure, wantError: worktree.ErrReadFailure},
		{name: "checkout failure", checkoutError: worktree.ErrCheckoutFailure, wantError: worktree.ErrCheckoutFailure, wantCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkout := &fakeCheckout{
				mainCheckout:  "/main",
				mainError:     test.mainError,
				checkoutError: test.checkoutError,
			}
			useCase := worktree.NewSwitchWorktree(&fakeLister{}, checkout)
			result, err := useCase.Switch(context.Background(), worktree.SwitchCommand{Branch: "feature"})
			if !errors.Is(err, test.wantError) || checkout.checkoutCalls != test.wantCalls {
				t.Fatalf("got %+v, %v, %d checkout calls", result, err, checkout.checkoutCalls)
			}
			if err == nil && (result.Path != checkout.mainCheckout || checkout.directory != checkout.mainCheckout || checkout.branch != "feature") {
				t.Fatalf("got %+v, checked out %q in %q", result, checkout.branch, checkout.directory)
			}
			if err != nil && result.Path != "" {
				t.Fatalf("failed checkout returned destination %q", result.Path)
			}
		})
	}
}
