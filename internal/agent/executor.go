package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type ExecutionRequest struct {
	ProyekID  domain.ID
	TugasID   *domain.ID
	AgenID    domain.ID
	Program   string
	Arguments []string
	Dir       string
	Env       []string
}

type ExecutionResult struct {
	Eksekusi domain.Eksekusi
}

type Executor struct {
	Runner Runner
	Repo   *storage.Repository
}

// Execute menjalankan program dalam konteks Proyek, Tugas, dan Agen.
// Riwayat Eksekusi tetap disimpan walaupun proses gagal atau dibatalkan.
func (e Executor) Execute(ctx context.Context, request ExecutionRequest) (ExecutionResult, error) {
	if e.Repo == nil {
		return ExecutionResult{}, errors.New("executor: repository wajib diisi")
	}
	if err := validateExecutionRequest(request); err != nil {
		return ExecutionResult{}, err
	}
	proyek, err := e.Repo.GetProyek(ctx, request.ProyekID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: proyek: %w", err)
	}
	agen, err := e.Repo.GetAgen(ctx, request.AgenID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: agen: %w", err)
	}
	if agen.RuangID != proyek.RuangID {
		return ExecutionResult{}, fmt.Errorf("executor: agen tidak termasuk ruang proyek: %w", storage.ErrInvalid)
	}

	if request.TugasID != nil {
		tugas, err := e.Repo.GetTugasNative(ctx, *request.TugasID)
		if err != nil {
			return ExecutionResult{}, fmt.Errorf("executor: tugas: %w", err)
		}
		if tugas.ProyekID != request.ProyekID || tugas.RuangID != proyek.RuangID {
			return ExecutionResult{}, fmt.Errorf("executor: tugas tidak termasuk proyek: %w", storage.ErrInvalid)
		}
		if !tugas.Status.CanTransitionTo(domain.StatusRunning) {
			return ExecutionResult{}, fmt.Errorf("executor: tugas berstatus %q tidak dapat dijalankan: %w", tugas.Status, storage.ErrInvalid)
		}
		assignments, assignmentErr := e.Repo.ListPenugasanByTugas(ctx, *request.TugasID)
		if assignmentErr != nil {
			return ExecutionResult{}, fmt.Errorf("executor: penugasan tugas: %w", assignmentErr)
		}
		if len(assignments) > 0 {
			assigned := false
			for _, assignment := range assignments {
				if assignment.AgenID == request.AgenID {
					assigned = true
					break
				}
			}
			if !assigned {
				return ExecutionResult{}, fmt.Errorf("executor: agen bukan pelaksana tugas: %w", storage.ErrInvalid)
			}
		}
	}

	approvalStatus, hasApproval, err := e.Repo.ApprovalGate(ctx, request.ProyekID, request.TugasID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: approval: %w", err)
	}
	if hasApproval {
		switch approvalStatus {
		case domain.ApprovalPending:
			return ExecutionResult{}, fmt.Errorf("executor: menunggu persetujuan: %w", storage.ErrInvalid)
		case domain.ApprovalRejected:
			return ExecutionResult{}, fmt.Errorf("executor: tindakan ditolak: %w", storage.ErrInvalid)
		case domain.ApprovalApproved:
			// lanjutkan eksekusi
		default:
			return ExecutionResult{}, fmt.Errorf("executor: status approval tidak dikenal: %w", storage.ErrInvalid)
		}
	}

	now := time.Now().UTC()
	eksekusi := domain.Eksekusi{
		ID:        newID(),
		RuangID:   proyek.RuangID,
		ProyekID:  request.ProyekID,
		TugasID:   request.TugasID,
		AgenID:    request.AgenID,
		Status:    domain.StatusRunning,
		Program:   request.Program,
		Arguments: append([]string(nil), request.Arguments...),
		StartedAt: now,
	}
	if err := e.Repo.CreateEksekusi(ctx, eksekusi); err != nil {
		return ExecutionResult{}, fmt.Errorf("executor: mulai eksekusi: %w", err)
	}
	e.catatPeristiwa(eksekusi, domain.EventExecutionStarted)

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
	persistErr := e.Repo.UpdateEksekusi(persistCtx, eksekusi.ID, status, exitCode, result.Stdout, result.Stderr, finished)

	eksekusi.Status = status
	eksekusi.ExitCode = exitCode
	eksekusi.Stdout = result.Stdout
	eksekusi.Stderr = result.Stderr
	eksekusi.FinishedAt = &finished

	if persistErr == nil {
		e.catatPeristiwa(eksekusi, domain.EventTypeForStatus(status))
	}

	return ExecutionResult{Eksekusi: eksekusi}, errors.Join(runErr, persistErr)
}

// catatPeristiwa menyimpan peristiwa Eksekusi sebagai jejak aktivitas lokal.
// Pencatatan bersifat best-effort: kegagalan menulis peristiwa tidak boleh
// menggagalkan Eksekusi, dan isi proses (argumen, stdout, stderr) tidak dicatat.
func (e Executor) catatPeristiwa(eksekusi domain.Eksekusi, jenis string) {
	if jenis == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id := eksekusi.ID
	_ = e.Repo.AppendEvent(ctx, domain.Event{
		RuangID:     eksekusi.RuangID,
		ExecutionID: &id,
		ProyekID:    eksekusi.ProyekID,
		TugasID:     eksekusi.TugasID,
		AgenID:      &eksekusi.AgenID,
		Type:        jenis,
		CreatedAt:   time.Now().UTC(),
	})
}

func validateExecutionRequest(request ExecutionRequest) error {
	if request.ProyekID == "" || request.AgenID == "" || strings.TrimSpace(request.Program) == "" {
		return fmt.Errorf("executor: %w", storage.ErrInvalid)
	}
	if request.Dir != "" && !filepath.IsAbs(request.Dir) {
		return fmt.Errorf("executor: direktori kerja harus berupa path absolut: %w", storage.ErrInvalid)
	}
	values := append([]string{request.Program, request.Dir}, request.Arguments...)
	values = append(values, request.Env...)
	for _, value := range values {
		if strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("executor: input proses mengandung karakter NUL: %w", storage.ErrInvalid)
		}
	}
	return nil
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
