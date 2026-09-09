package cliin

import (
	"bytes"
	"context"
	"testing"

	"github.com/ekkert-works/worktree/internal/worktree"
)

type stubSwitch struct{ err error }

func (s stubSwitch) Switch(context.Context, worktree.SwitchCommand) (worktree.SwitchResult, error) {
	return worktree.SwitchResult{Path: "/feature tree"}, s.err
}

func TestRun(t *testing.T) {
	for _, test := range []struct {
		name      string
		arguments []string
		err       error
		code      int
		output    string
	}{
		{name: "switch", arguments: []string{"switch", "feature"}, output: "/feature tree\n"},
		{name: "usage", code: 2},
		{name: "unknown command", arguments: []string{"add", "feature"}, code: 2},
		{name: "invalid branch", arguments: []string{"switch", ""}, err: worktree.ErrInvalidBranch, code: 2},
		{name: "not found", arguments: []string{"switch", "missing"}, err: worktree.ErrNotFound, code: 1},
		{name: "git failure", arguments: []string{"switch", "feature"}, err: worktree.ErrReadFailure, code: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), test.arguments, stubSwitch{err: test.err}, &stdout, &stderr)
			if code != test.code || stdout.String() != test.output {
				t.Fatalf("got code %d, output %q", code, stdout.String())
			}
			if (stderr.Len() > 0) != (test.code != 0) {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
		})
	}
}
