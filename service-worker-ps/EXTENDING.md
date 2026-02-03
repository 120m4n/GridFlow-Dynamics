# Extending to Other Databases

This document provides examples of how to extend the service worker to support other databases like Oracle and MongoDB.

## Repository Pattern Overview

The service worker uses the Repository pattern which provides an abstraction over the database layer. The key interface is:

```go
type Repository interface {
    Save(ctx context.Context, data *InventarioData) error
    Close() error
    HealthCheck(ctx context.Context) error
}
```

## Adding Oracle Support

### Step 1: Create Oracle Repository Implementation

Create `service-worker-ps/internal/repository/oracle.go`:

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/godror/godror" // Oracle driver
)

type OracleRepository struct {
    db *sql.DB
}

func NewOracleRepository(connectionString string) (*OracleRepository, error) {
    // Example connection string: "user/password@localhost:1521/ORCLPDB1"
    db, err := sql.Open("godror", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to open oracle connection: %w", err)
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
        return nil, fmt.Errorf("failed to ping oracle: %w", err)
    }

    return &OracleRepository{db: db}, nil
}

func (r *OracleRepository) Save(ctx context.Context, data *InventarioData) error {
    query := `
        INSERT INTO cuadrillas (
            grupo_trabajo, nombre_empleado, timestamp, latitud, longitud,
            codigo_odt, estado, porcentaje_progreso, nivel_bateria
        ) VALUES (:1, :2, :3, :4, :5, :6, :7, :8, :9)
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

func (r *OracleRepository) Close() error {
    if r.db != nil {
        return r.db.Close()
    }
    return nil
}

func (r *OracleRepository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
```

### Step 2: Update Factory

Update `service-worker-ps/internal/repository/factory.go`:

```go
func (f *Factory) CreateRepository(repoType RepositoryType, connectionString string) (Repository, error) {
    switch repoType {
    case PostgreSQL:
        return NewPostgresRepository(connectionString)
    case Oracle:
        return NewOracleRepository(connectionString)  // Add this line
    case MongoDB:
        return nil, fmt.Errorf("mongodb repository not implemented yet")
    default:
        return nil, fmt.Errorf("unknown repository type: %s", repoType)
    }
}
```

### Step 3: Update go.mod

Add Oracle driver to `service-worker-ps/go.mod`:

```bash
cd service-worker-ps
go get github.com/godror/godror
go mod tidy
```

### Step 4: Create Oracle Schema

Create `scripts/init_oracle.sql`:

```sql
-- Oracle schema for cuadrillas table
CREATE TABLE cuadrillas (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    grupo_trabajo VARCHAR2(255) NOT NULL,
    nombre_empleado VARCHAR2(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    latitud NUMBER(9,6) NOT NULL,
    longitud NUMBER(9,6) NOT NULL,
    codigo_odt VARCHAR2(255) NOT NULL,
    estado VARCHAR2(50) NOT NULL CHECK (estado IN ('en_ruta', 'trabajando', 'en_pausa', 'finalizado')),
    porcentaje_progreso NUMBER(3) NOT NULL CHECK (porcentaje_progreso >= 0 AND porcentaje_progreso <= 100),
    nivel_bateria NUMBER(3) NOT NULL CHECK (nivel_bateria >= 0 AND nivel_bateria <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_cuadrillas_grupo_trabajo ON cuadrillas(grupo_trabajo);
CREATE INDEX idx_cuadrillas_codigo_odt ON cuadrillas(codigo_odt);
CREATE INDEX idx_cuadrillas_timestamp ON cuadrillas(timestamp);
CREATE INDEX idx_cuadrillas_estado ON cuadrillas(estado);
```

### Step 5: Configure Environment

Update `.env` or environment variables:

```bash
REPOSITORY_TYPE=oracle
DATABASE_URL=user/password@localhost:1521/ORCLPDB1
```

## Adding MongoDB Support

### Step 1: Create MongoDB Repository Implementation

Create `service-worker-ps/internal/repository/mongodb.go`:

```go
package repository

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
    client     *mongo.Client
    collection *mongo.Collection
}

func NewMongoRepository(connectionString string) (*MongoRepository, error) {
    // Example connection string: "mongodb://user:pass@localhost:27017/gridflow"
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(connectionString)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
    }

    // Test connection
    if err := client.Ping(ctx, nil); err != nil {
        return nil, fmt.Errorf("failed to ping mongodb: %w", err)
    }

    // Get database and collection from connection string
    db := client.Database("gridflow")
    collection := db.Collection("cuadrillas")

    return &MongoRepository{
        client:     client,
        collection: collection,
    }, nil
}

func (r *MongoRepository) Save(ctx context.Context, data *InventarioData) error {
    doc := map[string]interface{}{
        "grupo_trabajo":       data.GrupoTrabajo,
        "nombre_empleado":     data.NombreEmpleado,
        "timestamp":           data.Timestamp,
        "latitud":             data.Latitud,
        "longitud":            data.Longitud,
        "codigo_odt":          data.CodigoODT,
        "estado":              data.Estado,
        "porcentaje_progreso": data.PorcentajeProgreso,
        "nivel_bateria":       data.NivelBateria,
        "created_at":          time.Now(),
    }

    _, err := r.collection.InsertOne(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to insert document: %w", err)
    }

    return nil
}

func (r *MongoRepository) Close() error {
    if r.client != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return r.client.Disconnect(ctx)
    }
    return nil
}

func (r *MongoRepository) HealthCheck(ctx context.Context) error {
    return r.client.Ping(ctx, nil)
}
```

### Step 2: Update Factory

Update `service-worker-ps/internal/repository/factory.go`:

```go
func (f *Factory) CreateRepository(repoType RepositoryType, connectionString string) (Repository, error) {
    switch repoType {
    case PostgreSQL:
        return NewPostgresRepository(connectionString)
    case Oracle:
        return NewOracleRepository(connectionString)
    case MongoDB:
        return NewMongoRepository(connectionString)  // Add this line
    default:
        return nil, fmt.Errorf("unknown repository type: %s", repoType)
    }
}
```

### Step 3: Update go.mod

Add MongoDB driver:

```bash
cd service-worker-ps
go get go.mongodb.org/mongo-driver/mongo
go mod tidy
```

### Step 4: Configure Environment

Update `.env` or environment variables:

```bash
REPOSITORY_TYPE=mongodb
DATABASE_URL=mongodb://gridflow_user:gridflow_password@localhost:27017/gridflow
```

### Step 5: Create Indexes in MongoDB

```javascript
// MongoDB shell or application code
use gridflow

db.cuadrillas.createIndex({ "grupo_trabajo": 1 })
db.cuadrillas.createIndex({ "codigo_odt": 1 })
db.cuadrillas.createIndex({ "timestamp": 1 })
db.cuadrillas.createIndex({ "estado": 1 })
```

## Docker Compose Integration

### For Oracle

Add to `docker-compose.yml`:

```yaml
  oracle:
    image: container-registry.oracle.com/database/express:21.3.0-xe
    container_name: gridflow-oracle
    environment:
      ORACLE_PWD: oracle_password
    ports:
      - "1521:1521"
    networks:
      - gridflow-network
    healthcheck:
      test: ["CMD", "sqlplus", "-L", "sys/oracle_password@//localhost:1521/XEPDB1 as sysdba", "@healthcheck.sql"]
      interval: 30s
      timeout: 10s
      retries: 5
```

### For MongoDB

Add to `docker-compose.yml`:

```yaml
  mongodb:
    image: mongo:7-jammy
    container_name: gridflow-mongodb
    environment:
      MONGO_INITDB_ROOT_USERNAME: gridflow_user
      MONGO_INITDB_ROOT_PASSWORD: gridflow_password
      MONGO_INITDB_DATABASE: gridflow
    ports:
      - "27017:27017"
    volumes:
      - mongodb-data:/data/db
    networks:
      - gridflow-network
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5
```

Don't forget to add volumes:

```yaml
volumes:
  postgres-data:
    driver: local
  mongodb-data:
    driver: local
```

## Summary

The extensible architecture allows switching between databases with these simple steps:

1. **Implement the Repository interface** for your database
2. **Update the Factory** to create instances of your repository
3. **Add database driver** to go.mod
4. **Update configuration** to use the new repository type
5. **Create database schema** (if needed)

No changes to the core worker pool, subscriber, or main application logic are required!
