package gitcli

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func TestParse(t *testing.T) {
	output := "worktree /bare\x00bare\x00\x00worktree /with spaces\nand newline\x00branch refs/heads/feature\x00\x00worktree /detached\x00detached\x00\x00"
	result, err := parse(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 || result[0].Path() != "/with spaces\nand newline" || !result[0].MatchesBranch("feature") {
		t.Fatalf("unexpected worktrees: %+v", result)
	}
	_, err = parse("branch refs/heads/feature\x00\x00")
	if !errors.Is(err, worktree.ErrReadFailure) {
		t.Fatalf("got %v", err)
	}
}

func TestListRealWorktrees(t *testing.T) {
	directory := t.TempDir()
	runGit := func(arguments ...string) {
		t.Helper()
		command := exec.Command("git", arguments...)
		command.Dir = directory
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, output)
		}
	}
	runGit("init", "--initial-branch=main")
	runGit("-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "initial")
	destination := filepath.Join(directory, "feature tree")
	runGit("worktree", "add", "-b", "feature", destination)
	result, err := worktree.NewSwitchWorktree(NewLister(directory), NewCheckout(directory), NewCheckout(directory)).Switch(context.Background(), worktree.SwitchCommand{Branch: "feature"})
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(destination)
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != want {
		t.Fatalf("got %q, want %q", result.Path, want)
	}
}

func TestTranslateProcessFailures(t *testing.T) {
	for _, test := range []struct {
		name       string
		missingGit bool
		want       error
	}{
		{name: "outside repository", want: worktree.ErrReadFailure},
		{name: "missing executable", missingGit: true, want: worktree.ErrUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.missingGit {
				t.Setenv("PATH", t.TempDir())
			}
			_, err := NewLister(t.TempDir()).List(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			var exitError *exec.ExitError
			var executableError *exec.Error
			if errors.As(err, &exitError) || errors.As(err, &executableError) || errors.Is(err, exec.ErrNotFound) {
				t.Fatalf("process error escaped: %v", err)
			}
		})
	}
}
