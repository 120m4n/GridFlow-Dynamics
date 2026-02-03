// Package repository provides a factory for creating repository instances.
package repository

import (
	"fmt"
)

// Factory creates repository instances based on configuration.
type Factory struct{}

// NewFactory creates a new repository factory.
func NewFactory() *Factory {
	return &Factory{}
}

// CreateRepository creates a repository instance based on the type and connection string.
// This allows easy switching between different database backends.
//
// Examples:
//   - PostgreSQL: CreateRepository(PostgreSQL, "postgres://user:pass@host:port/db?sslmode=disable")
//   - Oracle: CreateRepository(Oracle, "oracle://user:pass@host:port/service")
//   - MongoDB: CreateRepository(MongoDB, "mongodb://user:pass@host:port/db")
func (f *Factory) CreateRepository(repoType RepositoryType, connectionString string) (Repository, error) {
	switch repoType {
	case PostgreSQL:
		return NewPostgresRepository(connectionString)
	case Oracle:
		// Future implementation for Oracle
		return nil, fmt.Errorf("oracle repository not implemented yet")
	case MongoDB:
		// Future implementation for MongoDB
		return nil, fmt.Errorf("mongodb repository not implemented yet")
	default:
		return nil, fmt.Errorf("unknown repository type: %s", repoType)
	}
}
