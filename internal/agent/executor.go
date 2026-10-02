package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type ExecutionRequest struct {
	PekerjaanID domain.ID
	TugasID     *domain.ID
	AgenID      domain.ID
	Program     string
	Arguments   []string
	Dir         string
	Env         []string
}

type ExecutionResult struct {
	Sesi domain.Sesi
	Run  domain.Run
}

type Executor struct {
	Runner Runner
	Repo   *storage.Repository
}

// Execute menghubungkan satu proses dengan pekerjaan dan sesi agen.
// Riwayat run disimpan walaupun proses gagal atau dibatalkan.
func (e Executor) Execute(ctx context.Context, request ExecutionRequest) (ExecutionResult, error) {
	if e.Repo == nil {
		return ExecutionResult{}, errors.New("executor: repository wajib diisi")
	}
	if request.PekerjaanID == "" || request.AgenID == "" || request.Program == "" {
		return ExecutionResult{}, fmt.Errorf("executor: %w", storage.ErrInvalid)
	}
	if _, err := e.Repo.GetPekerjaan(ctx, request.PekerjaanID); err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: pekerjaan: %w", err)
	}
	if _, err := e.Repo.GetAgen(ctx, request.AgenID); err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: agen: %w", err)
	}

	now := time.Now().UTC()
	sesi := domain.Sesi{
		ID:          newID(),
		PekerjaanID: request.PekerjaanID,
		AgenID:      request.AgenID,
		Status:      domain.StatusRunning,
		StartedAt:   now,
	}
	if err := e.Repo.CreateSesi(ctx, sesi); err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: buat sesi: %w", err)
	}

	run := domain.Run{
		ID:          newID(),
		PekerjaanID: request.PekerjaanID,
		TugasID:     request.TugasID,
		AgenID:      request.AgenID,
		Status:      domain.StatusRunning,
		Program:     request.Program,
		Arguments:   append([]string(nil), request.Arguments...),
		StartedAt:   now,
	}
	if err := e.Repo.CreateRun(ctx, run); err != nil {
		_ = finishSesi(e.Repo, sesi.ID, domain.StatusFailed)
		return ExecutionResult{}, fmt.Errorf("executor: buat run: %w", err)
	}

	result, runErr := e.Runner.Run(ctx, Command{
		Program: request.Program,
		Args:    request.Arguments,
		Dir:     request.Dir,
		Env:     request.Env,
	})
	status := domain.StatusCompleted
	if runErr != nil {
		status = classifyStatus(runErr)
	}
	var exitCode *int
	if result.ExitCode >= 0 {
		value := result.ExitCode
		exitCode = &value
	}

	persistCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	finished := result.Finished
	if finished.IsZero() {
		finished = time.Now().UTC()
	}
	_ = e.Repo.FinishRun(persistCtx, run.ID, status, exitCode, result.Stdout, result.Stderr, finished)
	_ = e.Repo.FinishSesi(persistCtx, sesi.ID, status, finished)

	run.Status = status
	run.ExitCode = exitCode
	run.Stdout = result.Stdout
	run.Stderr = result.Stderr
	run.FinishedAt = &finished
	sesi.Status = status
	sesi.FinishedAt = &finished

	return ExecutionResult{Sesi: sesi, Run: run}, runErr
}

func classifyStatus(err error) domain.Status {
	switch {
	case errors.Is(err, ErrCancelled):
		return domain.StatusCancelled
	case errors.Is(err, ErrTimedOut):
		return domain.StatusFailed
	default:
		return domain.StatusFailed
	}
}

func finishSesi(repo *storage.Repository, id domain.ID, status domain.Status) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return repo.FinishSesi(ctx, id, status, time.Now().UTC())
}

func newID() domain.ID {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return domain.ID(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	return domain.ID(hex.EncodeToString(data[:]))
}
