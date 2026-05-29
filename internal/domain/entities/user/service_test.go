package user

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
	"gorm.io/gorm"
)

func TestWrapServiceErr_preservesSentinels(t *testing.T) {
	t.Parallel()

	repoErr := mapRepoErr("find by id", gorm.ErrRecordNotFound)
	wrapped := wrapServiceErr("get by id", repoErr)

	if !errors.Is(wrapped, ErrNotFound) {
		t.Fatalf("errors.Is(wrapped, ErrNotFound) = false, got: %v", wrapped)
	}
}

func TestWrapServiceErr_preservesDuplicateEmail(t *testing.T) {
	t.Parallel()

	repoErr := fmt.Errorf("user: create: %w", ErrDuplicateEmail)
	wrapped := wrapServiceErr("create", repoErr)

	if !errors.Is(wrapped, ErrDuplicateEmail) {
		t.Fatalf("errors.Is(wrapped, ErrDuplicateEmail) = false, got: %v", wrapped)
	}
}

func TestWrapServiceErr_nil(t *testing.T) {
	t.Parallel()

	if got := wrapServiceErr("test", nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestService_Update_nilName_returnsValidationError(t *testing.T) {
	t.Parallel()

	s := &Service{}
	_, err := s.Update(context.Background(), uuid.Nil, UserUpdate{Name: nil})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var valErr *validations.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	msg, ok := valErr.Fields["name"]
	if !ok {
		t.Fatalf("expected 'name' field in validation error, got %v", valErr.Fields)
	}
	if msg != "El nombre es requerido" {
		t.Fatalf("expected 'El nombre es requerido', got %q", msg)
	}
}
