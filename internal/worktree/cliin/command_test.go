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

type stubAdd struct{ err error }

func (s stubAdd) Add(context.Context, worktree.AddCommand) (worktree.AddResult, error) {
	return worktree.AddResult{Path: "/feature tree"}, s.err
}

type stubCompletion struct {
	command  worktree.CompleteCommand
	branches []string
	err      error
}

func (s *stubCompletion) Complete(_ context.Context, command worktree.CompleteCommand) (worktree.CompleteResult, error) {
	s.command = command
	return worktree.CompleteResult{Branches: s.branches}, s.err
}

type stubLister struct {
	worktrees []worktree.Worktree
	err       error
}

func (s stubLister) List(context.Context) ([]worktree.Worktree, error) {
	return s.worktrees, s.err
}

func TestList(t *testing.T) {
	trees := []worktree.Worktree{
		worktree.NewWorktree("/main", "main"),
		worktree.NewWorktree("/detached", ""),
	}
	for _, test := range []struct {
		name      string
		worktrees []worktree.Worktree
		err       error
		code      int
		output    string
	}{
		{name: "worktrees", worktrees: trees, output: "main\t/main\n\t/detached\n"},
		{name: "empty"},
		{name: "failure", err: worktree.ErrReadFailure, code: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), []string{"list"}, nil, nil, nil, stubLister{worktrees: test.worktrees, err: test.err}, &stdout, &stderr)
			if code != test.code || stdout.String() != test.output {
				t.Fatalf("got code %d, output %q", code, stdout.String())
			}
			if (stderr.Len() > 0) != (test.code != 0) {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
		})
	}
}

func TestCompletionProtocol(t *testing.T) {
	for _, test := range []struct {
		name      string
		arguments []string
		branches  []string
		err       error
		code      int
		output    string
	}{
		{name: "branches", arguments: []string{"__complete", "switch", "feat"}, branches: []string{"feature/one", "feature/two"}, output: "feature/one\nfeature/two\n"},
		{name: "empty", arguments: []string{"__complete", "switch", "feat"}},
		{name: "failure", arguments: []string{"__complete", "switch", "feat"}, err: worktree.ErrReadFailure, code: 1},
		{name: "invalid protocol", arguments: []string{"__complete"}, code: 2},
		{name: "unsupported command", arguments: []string{"__complete", "add", "feat"}, code: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			completion := &stubCompletion{branches: test.branches, err: test.err}
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), test.arguments, nil, nil, completion, nil, &stdout, &stderr)
			if code != test.code || stdout.String() != test.output || stderr.Len() != 0 {
				t.Fatalf("got code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
			}
			if test.code != 2 && completion.command.Prefix != "feat" {
				t.Fatalf("prefix not forwarded: %+v", completion.command)
			}
		})
	}
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
		{name: "add", arguments: []string{"add", "feature"}, output: "/feature tree\n"},
		{name: "usage", code: 2},
		{name: "unknown command", arguments: []string{"remove", "feature"}, code: 2},
		{name: "invalid branch", arguments: []string{"switch", ""}, err: worktree.ErrInvalidBranch, code: 2},
		{name: "add invalid branch", arguments: []string{"add", ""}, err: worktree.ErrInvalidBranch, code: 2},
		{name: "checkout failure", arguments: []string{"switch", "missing"}, err: worktree.ErrCheckoutFailure, code: 1},
		{name: "add failure", arguments: []string{"add", "missing"}, err: worktree.ErrAddFailure, code: 1},
		{name: "git failure", arguments: []string{"switch", "feature"}, err: worktree.ErrReadFailure, code: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), test.arguments, stubSwitch{err: test.err}, stubAdd{err: test.err}, nil, nil, &stdout, &stderr)
			if code != test.code || stdout.String() != test.output {
				t.Fatalf("got code %d, output %q", code, stdout.String())
			}
			if (stderr.Len() > 0) != (test.code != 0) {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
		})
	}
}
