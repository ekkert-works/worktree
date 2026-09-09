package shell_test

import (
	"os/exec"
	"testing"
)

// TestCompletion runs the shell completion script under both Bash and Zsh.
// The script exercises the real completion functions, so it must run in a
// shell rather than in Go. It expects the repository root as the working
// directory to build the binary and source shell/worktree.sh.
func TestCompletion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh"} {
		t.Run(shell, func(t *testing.T) {
			if _, err := exec.LookPath(shell); err != nil {
				t.Skipf("%s not installed", shell)
			}
			command := exec.Command(shell, "shell/completion_test.sh")
			command.Dir = ".."
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("%s: %v\n%s", shell, err, output)
			}
		})
	}
}
