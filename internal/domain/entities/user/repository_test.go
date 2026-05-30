package user

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestMapRepoErr(t *testing.T) {
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
			name:         "record not found maps to ErrNotFound",
			input:        gorm.ErrRecordNotFound,
			wantSentinel: ErrNotFound,
		},
		{
			name:         "duplicate key maps to ErrDuplicateEmail",
			input:        gorm.ErrDuplicatedKey,
			wantSentinel: ErrDuplicateEmail,
		},
		{
			name:         "unexpected DB error maps to ErrInternal",
			input:        errors.New("connection reset by peer"),
			wantSentinel: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRepoErr(context.Background(), "test action", tt.input)

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

func TestIsDuplicateKeyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   error
		want  bool
	}{
		{"nil", nil, false},
		{"gorm ErrDuplicatedKey", gorm.ErrDuplicatedKey, true},
		{"wrapped gorm ErrDuplicatedKey", fmt.Errorf("create: %w", gorm.ErrDuplicatedKey), true},
		{"pgconn PgError 23505", &pgconn.PgError{Code: "23505", Message: "duplicate key"}, true},
		{"wrapped pgconn PgError 23505", fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}), true},
		{"pgconn PgError other code", &pgconn.PgError{Code: "23503", Message: "foreign key"}, false},
		{"string contains duplicate key", errors.New("ERROR: duplicate key value violates unique constraint"), true},
		{"string contains unique constraint", errors.New("violates unique constraint \"users_email_key\""), true},
		{"generic DB error", errors.New("connection reset by peer"), false},
		{"record not found", gorm.ErrRecordNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicateKeyError(tt.err)
			if got != tt.want {
				t.Fatalf("isDuplicateKeyError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
