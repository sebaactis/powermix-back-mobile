package proof

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestMapProofRepoErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        error
		wantSentinel error
		wantNil      bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:         "unexpected DB error maps to ErrInternal",
			input:        errors.New("connection reset by peer"),
			wantSentinel: ErrInternal,
		},
		{
			name:         "record not found on list query maps to ErrInternal",
			input:        gorm.ErrRecordNotFound,
			wantSentinel: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapProofRepoErr("list proofs", tt.input)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}

			if !errors.Is(got, tt.wantSentinel) {
				t.Fatalf("errors.Is(got, %v) = false, got: %v", tt.wantSentinel, got)
			}
		})
	}
}
