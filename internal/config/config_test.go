package config

import (
	"os"
	"testing"
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
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("RABBITMQ_URL", "amqp://test:test@rabbitmq:5672/")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("RABBITMQ_URL")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := Load()

	if cfg.RabbitMQ.URL != "amqp://test:test@rabbitmq:5672/" {
		t.Errorf("Expected custom RabbitMQ URL, got %s", cfg.RabbitMQ.URL)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected custom server port 9090, got %s", cfg.Server.Port)
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
