package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg := Load()

	if cfg.RabbitMQ.URL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("Expected default RabbitMQ URL, got %s", cfg.RabbitMQ.URL)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default server port 8080, got %s", cfg.Server.Port)
	}

	if cfg.API.HMACSecret != "default-secret-change-in-production" {
		t.Errorf("Expected default HMAC secret, got %s", cfg.API.HMACSecret)
	}

	if cfg.API.RateLimitPerMin != 100 {
		t.Errorf("Expected default rate limit 100, got %d", cfg.API.RateLimitPerMin)
	}

	// Test idempotency defaults
	if !cfg.Idempotency.Enabled {
		t.Error("Expected idempotency to be enabled by default")
	}

	if cfg.Idempotency.TTL != 5*time.Minute {
		t.Errorf("Expected default idempotency TTL of 5m, got %v", cfg.Idempotency.TTL)
	}

	if cfg.Idempotency.CleanupInterval != time.Minute {
		t.Errorf("Expected default cleanup interval of 1m, got %v", cfg.Idempotency.CleanupInterval)
	}

	if cfg.Idempotency.Secret != "idempotency-secret-change-in-production" {
		t.Errorf("Expected default idempotency secret, got %s", cfg.Idempotency.Secret)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("RABBITMQ_URL", "amqp://test:test@rabbitmq:5672/")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("HMAC_SECRET", "custom-secret")
	defer func() {
		os.Unsetenv("RABBITMQ_URL")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("HMAC_SECRET")
	}()

	cfg := Load()

	if cfg.RabbitMQ.URL != "amqp://test:test@rabbitmq:5672/" {
		t.Errorf("Expected custom RabbitMQ URL, got %s", cfg.RabbitMQ.URL)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected custom server port 9090, got %s", cfg.Server.Port)
	}

	if cfg.API.HMACSecret != "custom-secret" {
		t.Errorf("Expected custom HMAC secret, got %s", cfg.API.HMACSecret)
	}
}

func TestLoadWithIdempotencyEnvVars(t *testing.T) {
	// Set idempotency environment variables
	os.Setenv("IDEMPOTENCY_ENABLED", "false")
	os.Setenv("IDEMPOTENCY_TTL", "10m")
	os.Setenv("IDEMPOTENCY_CLEANUP_INTERVAL", "2m")
	os.Setenv("IDEMPOTENCY_SECRET", "custom-idempotency-secret")
	defer func() {
		os.Unsetenv("IDEMPOTENCY_ENABLED")
		os.Unsetenv("IDEMPOTENCY_TTL")
		os.Unsetenv("IDEMPOTENCY_CLEANUP_INTERVAL")
		os.Unsetenv("IDEMPOTENCY_SECRET")
	}()

	cfg := Load()

	if cfg.Idempotency.Enabled {
		t.Error("Expected idempotency to be disabled")
	}

	if cfg.Idempotency.TTL != 10*time.Minute {
		t.Errorf("Expected idempotency TTL of 10m, got %v", cfg.Idempotency.TTL)
	}

	if cfg.Idempotency.CleanupInterval != 2*time.Minute {
		t.Errorf("Expected cleanup interval of 2m, got %v", cfg.Idempotency.CleanupInterval)
	}

	if cfg.Idempotency.Secret != "custom-idempotency-secret" {
		t.Errorf("Expected custom idempotency secret, got %s", cfg.Idempotency.Secret)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_KEY_SET",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s; want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns true when set to true",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns false when set to false",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns default when invalid value",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(%s, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_DURATION_NOT_SET",
			defaultValue: 5 * time.Minute,
			envValue:     "",
			expected:     5 * time.Minute,
		},
		{
			name:         "returns parsed duration",
			key:          "TEST_DURATION_SET",
			defaultValue: 5 * time.Minute,
			envValue:     "10m",
			expected:     10 * time.Minute,
		},
		{
			name:         "returns default when invalid value",
			key:          "TEST_DURATION_INVALID",
			defaultValue: 5 * time.Minute,
			envValue:     "invalid",
			expected:     5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvDuration(%s, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
