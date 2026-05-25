package jwtx

import (
	"testing"
)

// Test 4.2: Verify that jwt.NewJWT() returns error when JWT_SECRET is not set
// This validates the security constraint: no fallback to hardcoded "dev-secret"
func TestJWTNewJWT(t *testing.T) {
	t.Run("NewJWT returns error when JWT_SECRET is empty", func(t *testing.T) {
		// Clear JWT_SECRET
		t.Setenv("JWT_SECRET", "")
		// Set other required secrets
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "recovery_secret")

		jwt, err := NewJWT()

		if err == nil {
			t.Errorf("Expected error when JWT_SECRET is not set, got nil")
		}

		if jwt != nil {
			t.Errorf("Expected nil JWT when error occurs, got %+v", jwt)
		}

		// Verify the error message
		if err.Error() != "JWT_SECRET es requerido" {
			t.Errorf("Expected 'JWT_SECRET es requerido', got '%s'", err.Error())
		}
	})

	t.Run("NewJWT returns error when JWT_RECOVERY_PASS_SECRET is empty", func(t *testing.T) {
		// Set JWT_SECRET but clear JWT_RECOVERY_PASS_SECRET
		t.Setenv("JWT_SECRET", "main_secret")
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "")

		jwt, err := NewJWT()

		if err == nil {
			t.Errorf("Expected error when JWT_RECOVERY_PASS_SECRET is not set, got nil")
		}

		if jwt != nil {
			t.Errorf("Expected nil JWT when error occurs")
		}

		// Verify the error message
		if err.Error() != "JWT_RECOVERY_PASS_SECRET es requerido" {
			t.Errorf("Expected 'JWT_RECOVERY_PASS_SECRET es requerido', got '%s'", err.Error())
		}
	})

	t.Run("NewJWT returns valid JWT when both secrets are set", func(t *testing.T) {
		// Set both required secrets
		t.Setenv("JWT_SECRET", "main_secret_value")
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "recovery_secret_value")

		jwt, err := NewJWT()

		if err != nil {
			t.Errorf("Expected no error when secrets are set, got %v", err)
		}

		if jwt == nil {
			t.Errorf("Expected valid JWT instance when all secrets are set")
		}

		// Verify the JWT instance has the expected values
		if jwt.secret == nil || len(jwt.secret) == 0 {
			t.Errorf("Expected secret to be set in JWT instance")
		}

		if jwt.reset_secret == nil || len(jwt.reset_secret) == 0 {
			t.Errorf("Expected reset_secret to be set in JWT instance")
		}
	})

	t.Run("NewJWT sets default TTL values when no override env vars are set", func(t *testing.T) {
		// Set required secrets
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "reset_secret")
		// Clear any TTL overrides
		t.Setenv("JWT_TTL_MINUTES", "")
		t.Setenv("JWT_TTL_RECOVERY_MINUTES", "")
		t.Setenv("JWT_TTL_REFRESH_MINUTES", "")

		jwt, err := NewJWT()

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if jwt == nil {
			t.Fatalf("Expected valid JWT instance")
		}

		// Verify default TTLs
		expectedNormalTTL := 60    // minutes
		expectedResetTTL := 15     // minutes
		expectedRefreshTTL := 1440 // minutes (24 hours)

		if jwt.ttlNormal.Minutes() != float64(expectedNormalTTL) {
			t.Errorf("Expected normal TTL of %d minutes, got %f", expectedNormalTTL, jwt.ttlNormal.Minutes())
		}

		if jwt.ttlReset.Minutes() != float64(expectedResetTTL) {
			t.Errorf("Expected reset TTL of %d minutes, got %f", expectedResetTTL, jwt.ttlReset.Minutes())
		}

		if jwt.ttlRefresh.Minutes() != float64(expectedRefreshTTL) {
			t.Errorf("Expected refresh TTL of %d minutes, got %f", expectedRefreshTTL, jwt.ttlRefresh.Minutes())
		}
	})

	t.Run("NewJWT respects custom TTL values from env vars", func(t *testing.T) {
		// Set required secrets
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "reset_secret")
		// Set custom TTL values
		t.Setenv("JWT_TTL_MINUTES", "120")
		t.Setenv("JWT_TTL_RECOVERY_MINUTES", "30")
		t.Setenv("JWT_TTL_REFRESH_MINUTES", "2880")

		jwt, err := NewJWT()

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if jwt == nil {
			t.Fatalf("Expected valid JWT instance")
		}

		// Verify custom TTLs were applied
		if jwt.ttlNormal.Minutes() != 120 {
			t.Errorf("Expected normal TTL of 120 minutes, got %f", jwt.ttlNormal.Minutes())
		}

		if jwt.ttlReset.Minutes() != 30 {
			t.Errorf("Expected reset TTL of 30 minutes, got %f", jwt.ttlReset.Minutes())
		}

		if jwt.ttlRefresh.Minutes() != 2880 {
			t.Errorf("Expected refresh TTL of 2880 minutes, got %f", jwt.ttlRefresh.Minutes())
		}
	})

	t.Run("NewJWT ignores invalid TTL values and uses defaults", func(t *testing.T) {
		// Set required secrets
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("JWT_RECOVERY_PASS_SECRET", "reset_secret")
		// Set invalid TTL values (non-numeric or negative)
		t.Setenv("JWT_TTL_MINUTES", "invalid")
		t.Setenv("JWT_TTL_RECOVERY_MINUTES", "-10")

		jwt, err := NewJWT()

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if jwt == nil {
			t.Fatalf("Expected valid JWT instance")
		}

		// Verify defaults were used instead of invalid values
		if jwt.ttlNormal.Minutes() != 60 {
			t.Errorf("Expected default normal TTL of 60 minutes, got %f", jwt.ttlNormal.Minutes())
		}

		if jwt.ttlReset.Minutes() != 15 {
			t.Errorf("Expected default reset TTL of 15 minutes, got %f", jwt.ttlReset.Minutes())
		}
	})
}
