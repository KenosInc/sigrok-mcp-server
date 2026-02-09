package sigrok

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// TestHelperProcess is not a real test. It is used as a helper process
// by tests that need to mock exec.Command.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	exitCode, _ := strconv.Atoi(os.Getenv("GO_HELPER_EXIT_CODE"))
	stdout := os.Getenv("GO_HELPER_STDOUT")
	stderr := os.Getenv("GO_HELPER_STDERR")

	if stdout != "" {
		_, _ = fmt.Fprint(os.Stdout, stdout)
	}
	if stderr != "" {
		_, _ = fmt.Fprint(os.Stderr, stderr)
	}
	os.Exit(exitCode)
}

// fakeCommandFactory returns a CommandFactory that spawns the test binary
// as a helper process with the given stdout, stderr, and exit code.
func fakeCommandFactory(stdout, stderr string, exitCode int) CommandFactory {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"GO_HELPER_STDOUT="+stdout,
			"GO_HELPER_STDERR="+stderr,
			"GO_HELPER_EXIT_CODE="+strconv.Itoa(exitCode),
		)
		return cmd
	}
}

func TestExecutorRun(t *testing.T) {
	tests := []struct {
		name       string
		stdout     string
		stderr     string
		exitCode   int
		wantStdout string
		wantStderr string
		wantExit   int
		wantErr    bool
	}{
		{
			name:       "successful execution",
			stdout:     "sigrok-cli 0.7.2\n",
			stderr:     "",
			exitCode:   0,
			wantStdout: "sigrok-cli 0.7.2\n",
			wantStderr: "",
			wantExit:   0,
		},
		{
			name:       "non-zero exit code returns result not error",
			stdout:     "",
			stderr:     "Error: unknown option\n",
			exitCode:   1,
			wantStdout: "",
			wantStderr: "Error: unknown option\n",
			wantExit:   1,
		},
		{
			name:       "stdout and stderr combined",
			stdout:     "output data",
			stderr:     "warning message",
			exitCode:   0,
			wantStdout: "output data",
			wantStderr: "warning message",
			wantExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor("sigrok-cli", 30*time.Second, "")
			e.cmdFactory = fakeCommandFactory(tt.stdout, tt.stderr, tt.exitCode)

			result, err := e.Run(context.Background(), "--version")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Stdout != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", result.Stdout, tt.wantStdout)
			}
			if result.Stderr != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", result.Stderr, tt.wantStderr)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d", result.ExitCode, tt.wantExit)
			}
		})
	}
}

func TestExecutorRunBinaryNotFound(t *testing.T) {
	e := NewExecutor("/nonexistent/path/sigrok-cli", 30*time.Second, "")

	_, err := e.Run(context.Background(), "--version")
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestExecutorRunContextCancel(t *testing.T) {
	// Use a command factory that produces a long-running process.
	factory := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "60")
	}

	e := NewExecutor("sigrok-cli", 30*time.Second, "")
	e.cmdFactory = factory

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := e.Run(ctx, "--version")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestExecutorRunTimeout(t *testing.T) {
	// Use a very short timeout with a long-running command.
	factory := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Use a shell command that ignores signals and sleeps,
		// ensuring the process does not exit before the context deadline.
		return exec.CommandContext(ctx, "sleep", "60")
	}

	e := NewExecutor("sigrok-cli", 50*time.Millisecond, "")
	e.cmdFactory = factory

	_, err := e.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for timeout, got nil")
	}
}
