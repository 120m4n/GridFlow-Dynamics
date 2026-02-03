// Package repository provides PostgreSQL implementation of the Repository interface.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresRepository implements the Repository interface for PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository.
// connectionString format: "postgres://user:password@host:port/database?sslmode=disable"
func NewPostgresRepository(connectionString string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Save stores an inventory message in the PostgreSQL database.
func (r *PostgresRepository) Save(ctx context.Context, data *InventarioData) error {
	query := `
		INSERT INTO cuadrillas (
			grupo_trabajo, nombre_empleado, timestamp, latitud, longitud,
			codigo_odt, estado, porcentaje_progreso, nivel_bateria
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		data.GrupoTrabajo,
		data.NombreEmpleado,
		data.Timestamp,
		data.Latitud,
		data.Longitud,
		data.CodigoODT,
		data.Estado,
		data.PorcentajeProgreso,
		data.NivelBateria,
	)

	if err != nil {
		return fmt.Errorf("failed to insert into cuadrillas: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// HealthCheck verifies the database connection is healthy.
func (r *PostgresRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
