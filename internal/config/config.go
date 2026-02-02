// Package config provides configuration management for the GridFlow-Dynamics platform.
package config

import (
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	NATS   NATSConfig
	Server ServerConfig
	API    APIConfig
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
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

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		API: APIConfig{
			HMACSecret:      getEnv("HMAC_SECRET", "default-secret-change-in-production"),
			RateLimitPerMin: 100,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
