package gitcli

import (
	"os/exec"
	"strings"
	"testing"
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
