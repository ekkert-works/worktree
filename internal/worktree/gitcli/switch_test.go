package gitcli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

// switchBranch wires the switch use case over the real Git adapters: the lister
// resolves an existing worktree, and the adder creates one when none matches.
func switchBranch(directory, branch string) (worktree.SwitchResult, error) {
	adder := NewAdd(directory)
	useCase := worktree.NewSwitchWorktree(NewLister(directory), worktree.NewCreateWorktree(adder, adder))
	return useCase.Switch(context.Background(), worktree.SwitchCommand{Branch: branch})
}

func TestSwitchCreatesWorktreeWhenNoneMatches(t *testing.T) {
	directory := newRepository(t)
	gitOutput(t, directory, "branch", "existing")
	head := gitOutput(t, directory, "rev-parse", "HEAD")
	root, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		branch string
	}{
		{name: "existing branch is checked out", branch: "existing"},
		{name: "unknown name branches from the current directory", branch: "feature/new"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := switchBranch(directory, test.branch)
			if err != nil {
				t.Fatalf("switch %s: %v", test.branch, err)
			}
			want := root + ".worktrees/" + test.branch + "/" + filepath.Base(root)
			if result.Path != want {
				t.Fatalf("got %q, want %q", result.Path, want)
			}
			if branch := gitOutput(t, result.Path, "branch", "--show-current"); branch != test.branch {
				t.Fatalf("new worktree on %q, want %q", branch, test.branch)
			}
			// A new branch starts from the branch checked out in the current directory.
			if gitOutput(t, result.Path, "rev-parse", "HEAD") != head {
				t.Fatalf("new worktree not started from the current branch")
			}
		})
	}
}
