package gitcli

import (
	"context"
	"errors"
	"os/exec"
	"slices"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func TestCompleteLocalAndRemoteBranches(t *testing.T) {
	directory := newRepository(t)
	gitOutput(t, directory, "branch", "unused")
	gitOutput(t, directory, "remote", "add", "origin", directory)
	gitOutput(t, directory, "update-ref", "refs/remotes/origin/remote-only", "HEAD")
	gitOutput(t, directory, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/remote-only")
	lister := NewLister(directory)
	completion := worktree.NewCompleteBranches(lister, lister)
	result, err := completion.Complete(context.Background(), worktree.CompleteCommand{})
	want := []string{"main", "origin/remote-only", "remote-only", "unused"}
	if err != nil || !slices.Equal(result.Branches, want) {
		t.Fatalf("got %+v, %v; want %v", result, err, want)
	}
}

func TestBranchListerTranslatesFailures(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		missingGit bool
		want       error
	}{
		{name: "outside repository", want: worktree.ErrReadFailure},
		{name: "missing executable", missingGit: true, want: worktree.ErrUnavailable},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			directory := t.TempDir()
			if scenario.missingGit {
				t.Setenv("PATH", directory)
			}
			_, err := NewLister(directory).ListBranches(context.Background())
			var exitError *exec.ExitError
			var executableError *exec.Error
			if !errors.Is(err, scenario.want) || errors.As(err, &exitError) || errors.As(err, &executableError) {
				t.Fatalf("got %v, want translated %v", err, scenario.want)
			}
		})
	}
}
