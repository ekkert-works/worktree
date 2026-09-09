package worktree_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type fakeCheckout struct {
	location      worktree.CheckoutLocation
	locationError error
	checkoutError error
	locationCalls int
	checkoutCalls int
	branch        string
}

func (f *fakeCheckout) ReadLocation(context.Context) (worktree.CheckoutLocation, error) {
	f.locationCalls++
	return f.location, f.locationError
}

func (f *fakeCheckout) Checkout(_ context.Context, branch string) error {
	f.checkoutCalls++
	f.branch = branch
	return f.checkoutError
}

func TestCheckoutWithoutWorktree(t *testing.T) {
	for _, test := range []struct {
		name          string
		linked        bool
		locationError error
		checkoutError error
		wantError     error
		wantCalls     int
	}{
		{name: "main checkout", wantCalls: 1},
		{name: "linked worktree", linked: true, wantError: worktree.ErrLinkedCheckout},
		{name: "location failure", locationError: worktree.ErrReadFailure, wantError: worktree.ErrReadFailure},
		{name: "checkout failure", checkoutError: worktree.ErrCheckoutFailure, wantError: worktree.ErrCheckoutFailure, wantCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkout := &fakeCheckout{
				location:      worktree.CheckoutLocation{Path: "/main/subdirectory", Linked: test.linked},
				locationError: test.locationError,
				checkoutError: test.checkoutError,
			}
			useCase := worktree.NewSwitchWorktree(&fakeLister{}, checkout, checkout)
			result, err := useCase.Switch(context.Background(), worktree.SwitchCommand{Branch: "feature"})
			if !errors.Is(err, test.wantError) || checkout.checkoutCalls != test.wantCalls {
				t.Fatalf("got %+v, %v, %d checkout calls", result, err, checkout.checkoutCalls)
			}
			if err == nil && (result.Path != checkout.location.Path || checkout.branch != "feature") {
				t.Fatalf("got %+v, checked out %q", result, checkout.branch)
			}
			if err != nil && result.Path != "" {
				t.Fatalf("failed checkout returned destination %q", result.Path)
			}
		})
	}
}
