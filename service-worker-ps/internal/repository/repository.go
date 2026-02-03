// Package repository defines the repository interface for data persistence.
// This package uses the Repository pattern to allow easy switching between
// different database backends (PostgreSQL, Oracle, MongoDB, etc.)
package repository

import (
	"context"
	"time"
)

// InventarioData represents the data structure for inventory messages.
type InventarioData struct {
	GrupoTrabajo       string
	NombreEmpleado     string
	Timestamp          time.Time
	Latitud            float64
	Longitud           float64
	CodigoODT          string
	Estado             string
	PorcentajeProgreso int
	NivelBateria       int
}

// Repository defines the interface for data persistence operations.
// Implementations must provide methods for storing inventory data.
// This interface can be implemented for different databases:
// - PostgreSQL: PostgresRepository
// - Oracle: OracleRepository
// - MongoDB: MongoRepository
type Repository interface {
	// Save stores an inventory message in the database.
	Save(ctx context.Context, data *InventarioData) error

	// Close closes the repository and releases resources.
	Close() error

	// HealthCheck verifies the repository connection is healthy.
	HealthCheck(ctx context.Context) error
}

// RepositoryType represents the type of database repository.
type RepositoryType string

const (
	// PostgreSQL repository type
	PostgreSQL RepositoryType = "postgresql"
	// Oracle repository type (for future implementation)
	Oracle RepositoryType = "oracle"
	// MongoDB repository type (for future implementation)
	MongoDB RepositoryType = "mongodb"
)
