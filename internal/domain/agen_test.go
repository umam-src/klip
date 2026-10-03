package domain

import (
	"errors"
	"testing"
)

func TestAgenValidate(t *testing.T) {
	tests := []struct {
		name    string
		agent   Agen
		wantErr bool
	}{
		{
			name: "valid",
			agent: Agen{
				ID:     "agen-1",
				RuangID: "ruang-1",
				Name:   "Peneliti",
				Role:   "Peneliti produk",
				Status: AgenStatusActive,
			},
		},
		{
			name:    "missing id",
			agent:   Agen{RuangID: "ruang-1", Name: "Peneliti", Role: "Peneliti produk", Status: AgenStatusActive},
			wantErr: true,
		},
		{
			name:    "missing ruang",
			agent:   Agen{ID: "agen-1", Name: "Peneliti", Role: "Peneliti produk", Status: AgenStatusActive},
			wantErr: true,
		},
		{
			name:    "missing name",
			agent:   Agen{ID: "agen-1", RuangID: "ruang-1", Role: "Peneliti produk", Status: AgenStatusActive},
			wantErr: true,
		},
		{
			name:    "missing role",
			agent:   Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Peneliti", Status: AgenStatusActive},
			wantErr: true,
		},
		{
			name:    "unknown status",
			agent:   Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Peneliti", Role: "Peneliti produk", Status: "running"},
			wantErr: true,
		},
		{
			name: "self parent",
			agent: func() Agen {
				parent := ID("agen-1")
				return Agen{
					ID:       "agen-1",
					RuangID:  "ruang-1",
					Name:     "Peneliti",
					Role:     "Peneliti produk",
					Status:   AgenStatusActive,
					ParentID: &parent,
				}
			}(),
			wantErr: true,
		},
		{
			name: "inactive is valid",
			agent: Agen{
				ID:       "agen-1",
				RuangID:  "ruang-1",
				Name:     "Peneliti",
				Role:     "Peneliti produk",
				Status:   AgenStatusInactive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidAgen) {
					t.Fatalf("Validate() error = %v, want ErrInvalidAgen", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestAgenStatusIsKnown(t *testing.T) {
	if !AgenStatusActive.IsKnown() {
		t.Fatal("AgenStatusActive should be known")
	}
	if !AgenStatusInactive.IsKnown() {
		t.Fatal("AgenStatusInactive should be known")
	}
	if AgenStatus("running").IsKnown() {
		t.Fatal("runtime status should not be an Agen lifecycle status")
	}
}
