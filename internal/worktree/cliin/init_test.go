package cliin

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestInitShell(t *testing.T) {
	for _, test := range []struct {
		name      string
		arguments []string
		code      int
		contains  string
	}{
		{name: "bash", arguments: []string{"init", "bash"}, contains: "complete -F _worktree_complete_bash worktree"},
		{name: "zsh", arguments: []string{"init", "zsh"}, contains: "compdef _worktree_complete_zsh worktree"},
		{name: "missing shell", arguments: []string{"init"}, code: 2},
		{name: "unknown shell", arguments: []string{"init", "fish"}, code: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), test.arguments, nil, nil, &stdout, &stderr)
			if code != test.code {
				t.Fatalf("got code %d, want %d", code, test.code)
			}
			if test.code == 0 {
				if !strings.Contains(stdout.String(), "worktree() {") || !strings.Contains(stdout.String(), test.contains) {
					t.Fatalf("unexpected script: %q", stdout.String())
				}
				if stderr.Len() != 0 {
					t.Fatalf("unexpected stderr: %q", stderr.String())
				}
			} else if stderr.Len() == 0 {
				t.Fatalf("expected usage on stderr")
			}
		})
	}
}
