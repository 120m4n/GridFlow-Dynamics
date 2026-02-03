// Package config provides configuration management for the service worker.
package config

import (
	"os"
	"strconv"

	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/repository"
)

// Config holds all configuration for the service worker.
type Config struct {
	NATS       NATSConfig
	Database   DatabaseConfig
	Worker     WorkerConfig
	Repository repository.RepositoryType
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL     string
	Subject string
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	ConnectionString string
}

// WorkerConfig holds worker pool settings.
type WorkerConfig struct {
	NumWorkers     int
	BufferSize     int
	ShutdownTimeout int // seconds
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		NATS: NATSConfig{
			URL:     getEnv("NATS_URL", "nats://localhost:4222"),
			Subject: getEnv("NATS_SUBJECT", "inventario.cuadrilla"),
		},
		Database: DatabaseConfig{
			ConnectionString: getEnv(
				"DATABASE_URL",
				"postgres://gridflow_user:gridflow_password@localhost:5432/gridflow?sslmode=disable",
			),
		},
		Worker: WorkerConfig{
			NumWorkers:      getEnvInt("WORKER_NUM_WORKERS", 10),
			BufferSize:      getEnvInt("WORKER_BUFFER_SIZE", 100),
			ShutdownTimeout: getEnvInt("WORKER_SHUTDOWN_TIMEOUT", 30),
		},
		Repository: repository.RepositoryType(getEnv("REPOSITORY_TYPE", "postgresql")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
