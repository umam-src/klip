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

// Execute menghubungkan tugas dengan sesi agen dan riwayat run.
// Riwayat run tetap disimpan walaupun proses gagal atau dibatalkan.
func (e Executor) Execute(ctx context.Context, request ExecutionRequest) (ExecutionResult, error) {
	if e.Repo == nil {
		return ExecutionResult{}, errors.New("executor: repository wajib diisi")
	}
	if request.PekerjaanID == "" || request.AgenID == "" || request.Program == "" {
		return ExecutionResult{}, fmt.Errorf("executor: %w", storage.ErrInvalid)
	}
	pekerjaan, err := e.Repo.GetPekerjaan(ctx, request.PekerjaanID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: pekerjaan: %w", err)
	}
	agen, err := e.Repo.GetAgen(ctx, request.AgenID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: agen: %w", err)
	}
	if agen.RuangID != pekerjaan.RuangID {
		return ExecutionResult{}, fmt.Errorf("executor: agen tidak termasuk ruang pekerjaan: %w", storage.ErrInvalid)
	}
	if request.TugasID != nil {
		tugas, err := e.Repo.GetTugas(ctx, *request.TugasID)
		if err != nil {
			return ExecutionResult{}, fmt.Errorf("executor: tugas: %w", err)
		}
		if tugas.PekerjaanID != request.PekerjaanID {
			return ExecutionResult{}, fmt.Errorf("executor: tugas tidak termasuk pekerjaan: %w", storage.ErrInvalid)
		}
		if !tugas.Status.CanTransitionTo(domain.StatusRunning) {
			return ExecutionResult{}, fmt.Errorf("executor: tugas berstatus %q tidak dapat dijalankan: %w", tugas.Status, storage.ErrInvalid)
		}
	}

	now := time.Now().UTC()
	sesi := domain.Sesi{
		ID:          newID(),
		PekerjaanID: request.PekerjaanID,
		AgenID:      request.AgenID,
		Status:      domain.StatusRunning,
		StartedAt:   now,
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
	started, err := e.Repo.StartExecution(ctx, sesi, run)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: mulai execution: %w", err)
	}
	sesi = started.Sesi
	run = started.Run

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

	persistErr := e.Repo.FinalizeExecution(persistCtx, run.ID, sesi.ID, request.TugasID, status, exitCode, result.Stdout, result.Stderr, finished)

	run.Status = status
	run.ExitCode = exitCode
	run.Stdout = result.Stdout
	run.Stderr = result.Stderr
	run.FinishedAt = &finished
	sesi.Status = status
	sesi.FinishedAt = &finished

	return ExecutionResult{Sesi: sesi, Run: run}, errors.Join(runErr, persistErr)
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

func newID() domain.ID {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return domain.ID(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	return domain.ID(hex.EncodeToString(data[:]))
}
