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
			name:  "valid",
			agent: Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Peneliti"},
		},
		{
			name:    "missing id",
			agent:   Agen{RuangID: "ruang-1", Name: "Peneliti"},
			wantErr: true,
		},
		{
			name:    "missing ruang",
			agent:   Agen{ID: "agen-1", Name: "Peneliti"},
			wantErr: true,
		},
		{
			name:    "missing name",
			agent:   Agen{ID: "agen-1", RuangID: "ruang-1"},
			wantErr: true,
		},
		{
			name: "self parent",
			agent: func() Agen {
				parent := ID("agen-1")
				return Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Peneliti", ParentID: &parent}
			}(),
			wantErr: true,
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
