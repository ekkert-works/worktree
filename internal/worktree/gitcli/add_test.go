package gitcli

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func TestAddCreatesWorktreeForExistingBranch(t *testing.T) {
	directory := newRepository(t)
	gitOutput(t, directory, "branch", "feature/x")
	adder := NewAdd(directory)

	root, err := adder.ReadMainCheckout(context.Background())
	want, _ := filepath.EvalSymlinks(directory)
	if err != nil || root != want {
		t.Fatalf("main checkout: got %q, %v; want %q", root, err, want)
	}

	destination := root + ".worktrees/feature/x"
	if err := adder.Add(context.Background(), destination, "feature/x"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if head := gitOutput(t, destination, "branch", "--show-current"); head != "feature/x" {
		t.Fatalf("got branch %q in new worktree", head)
	}

	// A branch already checked out cannot be added again.
	if err := adder.Add(context.Background(), root+".worktrees/again", "feature/x"); !errors.Is(err, worktree.ErrAddFailure) {
		t.Fatalf("re-add: got %v, want ErrAddFailure", err)
	}
}

func TestAddAdapterTranslatesFailures(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		missingGit bool
	}{
		{name: "outside repository"},
		{name: "missing executable", missingGit: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			directory := t.TempDir()
			if scenario.missingGit {
				t.Setenv("PATH", directory)
			}
			adder := NewAdd(directory)
			_, rootError := adder.ReadMainCheckout(context.Background())
			addError := adder.Add(context.Background(), filepath.Join(directory, "tree"), "main")
			for _, test := range []struct{ err, want error }{
				{rootError, worktree.ErrReadFailure},
				{addError, worktree.ErrAddFailure},
			} {
				want := test.want
				if scenario.missingGit {
					want = worktree.ErrUnavailable
				}
				var exitError *exec.ExitError
				var executableError *exec.Error
				if !errors.Is(test.err, want) || errors.As(test.err, &exitError) || errors.As(test.err, &executableError) {
					t.Fatalf("got %v, want translated %v", test.err, want)
				}
			}
		})
	}
}
