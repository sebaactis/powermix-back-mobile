package token

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestMapTokenRepoErr(t *testing.T) {
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
			name:         "record not found maps to ErrTokenNotFound",
			input:        gorm.ErrRecordNotFound,
			wantSentinel: ErrTokenNotFound,
		},
		{
			name:         "unexpected DB error maps to ErrInternal",
			input:        errors.New("connection reset by peer"),
			wantSentinel: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapTokenRepoErr(context.Background(), "get by hash", tt.input)

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

func TestMapResetTokenRepoErr(t *testing.T) {
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
			name:         "record not found maps to ErrTokenInvalid",
			input:        gorm.ErrRecordNotFound,
			wantSentinel: ErrTokenInvalid,
		},
		{
			name:         "unexpected DB error maps to ErrInternal",
			input:        errors.New("connection reset by peer"),
			wantSentinel: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapResetTokenRepoErr(context.Background(), "get valid reset token", tt.input)

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
