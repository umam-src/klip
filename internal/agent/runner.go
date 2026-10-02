package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultMaxOutputBytes int64 = 1 << 20

var (
	ErrInvalidCommand = errors.New("perintah tidak valid")
	ErrTimedOut       = errors.New("proses melewati batas waktu")
	ErrCancelled      = errors.New("proses dibatalkan")
)

type Command struct {
	Program string
	Args    []string
	Dir     string
	Env     []string
}

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Started  time.Time
	Finished time.Time
}

type Runner struct {
	MaxOutputBytes int64
}

func (r Runner) Run(ctx context.Context, command Command) (Result, error) {
	if strings.TrimSpace(command.Program) == "" {
		return Result{}, ErrInvalidCommand
	}

	started := time.Now().UTC()
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = command.Env

	maxOutput := r.MaxOutputBytes
	if maxOutput <= 0 {
		maxOutput = defaultMaxOutputBytes
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{dst: &stdout, max: maxOutput}
	cmd.Stderr = &limitedWriter{dst: &stderr, max: maxOutput}

	err := cmd.Run()
	finished := time.Now().UTC()
	result := Result{Started: started, Finished: finished}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return result, fmt.Errorf("%w: %v", ErrTimedOut, ctx.Err())
		}
		return result, fmt.Errorf("%w: %v", ErrCancelled, ctx.Err())
	}
	if err != nil {
		return result, fmt.Errorf("jalankan proses: %w", err)
	}
	return result, nil
}

type limitedWriter struct {
	dst *bytes.Buffer
	max int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := w.max - int64(w.dst.Len())
	if remaining <= 0 {
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		_, _ = w.dst.Write(p[:remaining])
		return len(p), nil
	}
	return w.dst.Write(p)
}
