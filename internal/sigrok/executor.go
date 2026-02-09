package sigrok

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// CommandFactory creates an *exec.Cmd. It exists as a seam for testing.
type CommandFactory func(ctx context.Context, name string, args ...string) *exec.Cmd

// CommandResult holds the output of a sigrok-cli invocation.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Executor runs sigrok-cli commands.
type Executor struct {
	cliPath    string
	timeout    time.Duration
	workingDir string
	cmdFactory CommandFactory
}

// NewExecutor creates a new Executor with the given configuration.
func NewExecutor(cliPath string, timeout time.Duration, workingDir string) *Executor {
	return &Executor{
		cliPath:    cliPath,
		timeout:    timeout,
		workingDir: workingDir,
		cmdFactory: defaultCommandFactory,
	}
}

func defaultCommandFactory(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// Run executes sigrok-cli with the given arguments and returns the result.
// Non-zero exit codes are returned in CommandResult, not as errors.
// Errors are returned only for execution failures (binary not found, timeout, etc.).
func (e *Executor) Run(ctx context.Context, args ...string) (*CommandResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := e.cmdFactory(ctx, e.cliPath, args...)
	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If the context was cancelled or timed out, return that as an error
		// rather than treating it as a normal non-zero exit.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &CommandResult{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: exitErr.ExitCode(),
			}, nil
		}
		return nil, err
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}, nil
}
