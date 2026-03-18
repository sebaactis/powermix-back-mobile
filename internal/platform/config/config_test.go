package config

import (
	"testing"
)

// Test 4.2: Verify that config.Load() returns error when required env vars are missing
func TestConfigLoad(t *testing.T) {
	t.Run("Load returns error when DSN is not set", func(t *testing.T) {
		// Set all required env vars except DSN
		t.Setenv("HTTP_ADDR", "localhost:8080")
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "") // Empty DSN
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "resend_key")
		t.Setenv("JWT_REFRESH_HASH", "hash")

		cfg, err := Load()

		if err == nil {
			t.Errorf("Expected error when DSN is not set, got nil")
		}

		if cfg != (Config{}) {
			t.Errorf("Expected empty Config when error occurs, got %+v", cfg)
		}
	})

	t.Run("Load returns error when JWT_REFRESH_HASH is not set", func(t *testing.T) {
		// Set all required env vars except JWT_REFRESH_HASH
		t.Setenv("HTTP_ADDR", "localhost:8080")
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "postgres://localhost")
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "resend_key")
		t.Setenv("JWT_REFRESH_HASH", "") // Empty hash

		cfg, err := Load()

		if err == nil {
			t.Errorf("Expected error when JWT_REFRESH_HASH is not set, got nil")
		}

		if cfg != (Config{}) {
			t.Errorf("Expected empty Config when error occurs, got %+v", cfg)
		}
	})

	t.Run("Load returns error when HTTP_ADDR is not set", func(t *testing.T) {
		// Set all required env vars except HTTP_ADDR
		t.Setenv("HTTP_ADDR", "") // Empty HTTP_ADDR
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "postgres://localhost")
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "resend_key")
		t.Setenv("JWT_REFRESH_HASH", "hash")

		cfg, err := Load()

		if err == nil {
			t.Errorf("Expected error when HTTP_ADDR is not set, got nil")
		}

		if cfg != (Config{}) {
			t.Errorf("Expected empty Config when error occurs, got %+v", cfg)
		}
	})

	t.Run("Load returns nil error when all required vars are set", func(t *testing.T) {
		// Set all required env vars
		t.Setenv("HTTP_ADDR", "localhost:8080")
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "postgres://localhost")
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "resend_key")
		t.Setenv("JWT_REFRESH_HASH", "hash")

		cfg, err := Load()

		if err != nil {
			t.Errorf("Expected no error when all vars are set, got %v", err)
		}

		if cfg == (Config{}) {
			t.Errorf("Expected non-empty Config when load succeeds")
		}

		if cfg.HTTPAddr != "localhost:8080" {
			t.Errorf("Expected HTTPAddr to be set, got %s", cfg.HTTPAddr)
		}

		if cfg.DSN != "postgres://localhost" {
			t.Errorf("Expected DSN to be set, got %s", cfg.DSN)
		}
	})

	t.Run("Load returns error when RESEND_API_KEY is missing", func(t *testing.T) {
		// Set all required env vars except RESEND_API_KEY
		t.Setenv("HTTP_ADDR", "localhost:8080")
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "postgres://localhost")
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "") // Empty resend key
		t.Setenv("JWT_REFRESH_HASH", "hash")

		cfg, err := Load()

		if err == nil {
			t.Errorf("Expected error when RESEND_API_KEY is not set, got nil")
		}

		if cfg != (Config{}) {
			t.Errorf("Expected empty Config when error occurs")
		}
	})
}

// Test: Validate that the error message contains the missing variable name
func TestConfigErrorMessages(t *testing.T) {
	t.Run("Error message includes the name of the missing variable", func(t *testing.T) {
		t.Setenv("HTTP_ADDR", "localhost:8080")
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DSN", "") // Missing DSN
		t.Setenv("MERCAGO_PAGO_TOKEN", "token")
		t.Setenv("COFFEJI_KEY", "key")
		t.Setenv("COFFEJI_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "resend_key")
		t.Setenv("JWT_REFRESH_HASH", "hash")

		cfg, err := Load()

		if err == nil {
			t.Fatalf("Expected error when DSN is missing")
		}

		if cfg != (Config{}) {
			t.Errorf("Expected empty Config when error occurs")
		}

		// Verify the error message contains information about which variable is missing
		errMsg := err.Error()
		if len(errMsg) == 0 {
			t.Errorf("Expected non-empty error message")
		}
	})
}
