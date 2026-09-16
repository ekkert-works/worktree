package gitcli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func gitOutput(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return strings.TrimSuffix(string(output), "\n")
}

func newRepository(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	gitOutput(t, directory, "init", "--initial-branch=main")
	gitOutput(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "initial")
	return directory
}

func switchBranch(directory, branch string) (worktree.SwitchResult, error) {
	checkout := NewCheckout(directory)
	useCase := worktree.NewSwitchWorktree(NewLister(directory), checkout)
	return useCase.Switch(context.Background(), worktree.SwitchCommand{Branch: branch})
}

func TestCheckoutLocalAndRemoteBranches(t *testing.T) {
	directory := newRepository(t)
	gitOutput(t, directory, "branch", "unused")
	gitOutput(t, directory, "remote", "add", "origin", directory)
	gitOutput(t, directory, "update-ref", "refs/remotes/origin/remote-only", "HEAD")
	gitOutput(t, directory, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/remote-only")
	want, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, branch := range []string{"unused", "remote-only", "origin/remote-only"} {
		destination, err := switchBranch(directory, branch)
		if err != nil || destination.Path != want {
			t.Fatalf("switch %s: %+v, %v", branch, destination, err)
		}
		if branch == "origin/remote-only" {
			if head := gitOutput(t, directory, "rev-parse", "--abbrev-ref", "HEAD"); head != "HEAD" {
				t.Fatalf("qualified remote did not detach HEAD: %s", head)
			}
			continue
		}
		if head := gitOutput(t, directory, "branch", "--show-current"); head != branch {
			t.Fatalf("got branch %q, want %q", head, branch)
		}
	}
	if upstream := gitOutput(t, directory, "rev-parse", "--abbrev-ref", "remote-only@{upstream}"); upstream != "origin/remote-only" {
		t.Fatalf("got upstream %q", upstream)
	}
}

func TestLinkedWorktreeChecksOutInMainCheckout(t *testing.T) {
	directory := newRepository(t)
	gitOutput(t, directory, "branch", "unused")
	want, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, detached := range []bool{false, true} {
		name := "attached"
		arguments := []string{"worktree", "add", "-b", "linked"}
		if detached {
			name = "detached"
			arguments = []string{"worktree", "add", "--detach"}
		}
		linked := filepath.Join(t.TempDir(), name)
		gitOutput(t, directory, append(arguments, linked)...)
		subdirectory := filepath.Join(linked, "subdirectory")
		if err := os.Mkdir(subdirectory, 0755); err != nil {
			t.Fatal(err)
		}
		// A branch with no worktree is checked out in the main checkout, and the
		// linked worktree stays on its own branch.
		gitOutput(t, directory, "checkout", "main", "--")
		linkedHeadBefore := gitOutput(t, linked, "rev-parse", "--abbrev-ref", "HEAD")
		result, err := switchBranch(subdirectory, "unused")
		if err != nil || result.Path != want {
			t.Fatalf("%s: switch to unused: %+v, %v", name, result, err)
		}
		if head := gitOutput(t, directory, "branch", "--show-current"); head != "unused" {
			t.Fatalf("%s: main checkout on %q, want unused", name, head)
		}
		if headAfter := gitOutput(t, linked, "rev-parse", "--abbrev-ref", "HEAD"); headAfter != linkedHeadBefore {
			t.Fatalf("%s: linked worktree changed from %s to %s", name, linkedHeadBefore, headAfter)
		}
	}
}

func TestCheckoutDoesNotRestoreFilesOrOverwriteChanges(t *testing.T) {
	directory := newRepository(t)
	file := filepath.Join(directory, "tracked")
	if err := os.WriteFile(file, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, directory, "add", "tracked")
	gitOutput(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "commit", "-m", "add file")
	gitOutput(t, directory, "branch", "with-file")
	gitOutput(t, directory, "checkout", "main~1", "--")
	if err := os.WriteFile(file, []byte("local changes"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, branch := range []string{"with-file", "tracked", "missing", "--force"} {
		_, err := switchBranch(directory, branch)
		want := worktree.ErrCheckoutFailure
		if branch == "--force" {
			want = worktree.ErrInvalidBranch
		}
		if !errors.Is(err, want) {
			t.Fatalf("switch %s: got %v, want %v", branch, err, want)
		}
		content, err := os.ReadFile(file)
		if err != nil || string(content) != "local changes" {
			t.Fatalf("local changes overwritten: %q, %v", content, err)
		}
	}
}

func TestCheckoutAdapterTranslatesFailures(t *testing.T) {
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
			_, mainError := NewCheckout(directory).MainCheckout(context.Background())
			checkoutError := NewCheckout(directory).Checkout(context.Background(), directory, "main")
			for _, test := range []struct{ err, want error }{
				{mainError, worktree.ErrReadFailure},
				{checkoutError, worktree.ErrCheckoutFailure},
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
