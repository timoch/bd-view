package data

import (
	"context"
	"fmt"
	"os/exec"
)

// CommandExecutor defines the interface for running bd CLI commands.
// This enables test mocking by replacing the real executor with a stub.
type CommandExecutor interface {
	// Execute runs a bd command with the given arguments and returns stdout.
	Execute(ctx context.Context, args ...string) ([]byte, error)
}

// BdExecutor runs bd commands as subprocesses.
type BdExecutor struct {
	// BdPath is the path to the bd binary. Defaults to "bd".
	BdPath string
}

func (e *BdExecutor) Execute(ctx context.Context, args ...string) ([]byte, error) {
	bdPath := e.BdPath
	if bdPath == "" {
		bdPath = "bd"
	}

	cmd := exec.CommandContext(ctx, bdPath, args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd command failed: %s: %s", exitErr, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd command failed: %w", err)
	}
	return out, nil
}
