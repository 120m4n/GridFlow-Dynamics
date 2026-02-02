// Package config provides configuration management for the GridFlow-Dynamics platform.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	RabbitMQ    RabbitMQConfig
	Server      ServerConfig
	API         APIConfig
	Idempotency IdempotencyConfig
}

// RabbitMQConfig holds RabbitMQ connection settings.
type RabbitMQConfig struct {
	URL string
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Port string
}

// APIConfig holds API settings.
type APIConfig struct {
	HMACSecret      string
	RateLimitPerMin int
}

// IdempotencyConfig holds idempotency settings.
// These settings control how duplicate requests are detected and cached.
//
// For future Redis integration, add a "RedisURL" field and use it to
// initialize a Redis-based IdempotencyStore instead of the in-memory store.
type IdempotencyConfig struct {
	// Enabled controls whether idempotency checking is active.
	Enabled bool
	// TTL specifies how long idempotency records are kept.
	// After this duration, the same request can be processed again.
	TTL time.Duration
	// CleanupInterval specifies how often expired entries are removed.
	// Only applicable for in-memory store; Redis handles TTL automatically.
	CleanupInterval time.Duration
	// Secret is used for HMAC key generation (idempotency key).
	// Should be different from the HMAC signature secret for security.
	Secret string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		API: APIConfig{
			HMACSecret:      getEnv("HMAC_SECRET", "default-secret-change-in-production"),
			RateLimitPerMin: 100,
		},
		Idempotency: IdempotencyConfig{
			Enabled:         getEnvBool("IDEMPOTENCY_ENABLED", true),
			TTL:             getEnvDuration("IDEMPOTENCY_TTL", 5*time.Minute),
			CleanupInterval: getEnvDuration("IDEMPOTENCY_CLEANUP_INTERVAL", time.Minute),
			Secret:          getEnv("IDEMPOTENCY_SECRET", "idempotency-secret-change-in-production"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
