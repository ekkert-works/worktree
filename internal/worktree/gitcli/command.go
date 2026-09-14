package gitcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ekkert-works/worktree/internal/worktree"
)

func runGit(ctx context.Context, directory string, failure error, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", arguments...)
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if errors.Is(err, exec.ErrNotFound) {
		return "", worktree.ErrUnavailable
	}
	if err != nil {
		// Surface git's own reason, e.g. "already used by worktree at ...".
		if reason := strings.TrimSpace(stderr.String()); reason != "" {
			return "", fmt.Errorf("%w: %s", failure, reason)
		}
		return "", failure
	}
	return stdout.String(), nil
}
