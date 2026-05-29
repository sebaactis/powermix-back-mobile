package prode

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestMapProdeMatchErr(t *testing.T) {
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
			name:         "record not found maps to ErrMatchNotFound",
			input:        gorm.ErrRecordNotFound,
			wantSentinel: ErrMatchNotFound,
		},
		{
			name:         "unexpected DB error maps to ErrInternal",
			input:        errors.New("connection reset by peer"),
			wantSentinel: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapProdeMatchErr("get match", tt.input)

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

func TestMapProdeRepoErr(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapProdeRepoErr("list matches", tt.input)

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

func TestMapProdeUserPredictionLookup(t *testing.T) {
	t.Parallel()

	if got := mapProdeUserPredictionLookupErr(gorm.ErrRecordNotFound); got != nil {
		t.Fatalf("got %v, want nil for missing prediction", got)
	}

	if got := mapProdeUserPredictionLookupErr(errors.New("connection reset by peer")); !errors.Is(got, ErrInternal) {
		t.Fatalf("errors.Is(got, ErrInternal) = false, got: %v", got)
	}
}
