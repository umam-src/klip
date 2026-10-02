package agent

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunnerRun(t *testing.T) {
	program, args := testCommand("hello")
	result, err := (Runner{}).Run(context.Background(), Command{Program: program, Args: args})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d", result.ExitCode)
	}
	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Fatalf("stdout = %q", result.Stdout)
	}
	if result.Finished.Before(result.Started) {
		t.Fatalf("finished before started")
	}
}

func TestRunnerRejectsEmptyProgram(t *testing.T) {
	_, err := (Runner{}).Run(context.Background(), Command{})
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("error = %v, want ErrInvalidCommand", err)
	}
}

func TestRunnerClassifiesExitFailure(t *testing.T) {
	program, args := testCommand("echo-error")
	result, err := (Runner{}).Run(context.Background(), Command{Program: program, Args: args})
	if !errors.Is(err, ErrProcessFailed) {
		t.Fatalf("error = %v, want ErrProcessFailed", err)
	}
	if result.ExitCode == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(result.Stderr, "echo-error") {
		t.Fatalf("stderr = %q", result.Stderr)
	}
}

func TestRunnerClassifiesStartFailure(t *testing.T) {
	_, err := (Runner{}).Run(context.Background(), Command{Program: "klip-program-tidak-ada"})
	if !errors.Is(err, ErrStartFailed) {
		t.Fatalf("error = %v, want ErrStartFailed", err)
	}
}

func TestRunnerTimeout(t *testing.T) {
	program, args := testCommand("sleep")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := (Runner{}).Run(ctx, Command{Program: program, Args: args})
	if !errors.Is(err, ErrTimedOut) {
		t.Fatalf("error = %v, want ErrTimedOut", err)
	}
}

func TestRunnerLimitsOutput(t *testing.T) {
	program, args := testCommand("output")
	result, err := (Runner{MaxOutputBytes: 4}).Run(context.Background(), Command{Program: program, Args: args})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Stdout) != 4 {
		t.Fatalf("stdout length = %d, want 4", len(result.Stdout))
	}
}

func TestRunnerConcurrencyLimit(t *testing.T) {
	program, args := testCommand("sleep")
	runner := NewRunner(1, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := runner.Run(ctx, Command{Program: program, Args: args})
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	}
}

func testCommand(value string) (string, []string) {
	if runtime.GOOS == "windows" {
		switch value {
		case "hello":
			return "cmd", []string{"/C", "echo", "hello"}
		case "echo-error":
			return "cmd", []string{"/C", "echo", "echo-error", "1>&2", "&", "exit", "1"}
		case "sleep":
			return "powershell", []string{"-NoProfile", "-Command", "Start-Sleep -Milliseconds 500"}
		case "output":
			return "cmd", []string{"/C", "echo", "123456789"}
		}
	}

	switch value {
	case "hello":
		return "printf", []string{"hello\n"}
	case "echo-error":
		return "sh", []string{"-c", "printf echo-error >&2; exit 1"}
	case "sleep":
		return "sleep", []string{"1"}
	case "output":
		return "printf", []string{"123456789\n"}
	default:
		return "false", nil
	}
}
