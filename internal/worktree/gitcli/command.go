package gitcli

import (
	"context"
	"errors"
	"os/exec"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func runGit(ctx context.Context, directory string, failure error, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", arguments...)
	command.Dir = directory
	output, err := command.Output()
	if errors.Is(err, exec.ErrNotFound) {
		return "", worktree.ErrUnavailable
	}
	if err != nil {
		return "", failure
	}
	return string(output), nil
}
