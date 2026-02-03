# Service Worker PostgreSQL

Service worker independiente para GridFlow-Dynamics que escucha eventos de NATS y los persiste en PostgreSQL.

## Arquitectura

Este servicio implementa un **patrón Repository** extensible que permite cambiar fácilmente entre diferentes bases de datos (PostgreSQL, Oracle, MongoDB) con mínimos cambios de código.

### Componentes Principales

1. **Repository Pattern**: Abstracción de la capa de persistencia
   - `Repository` interface: Define las operaciones de persistencia
   - `PostgresRepository`: Implementación para PostgreSQL
   - `Factory`: Patrón Factory para crear instancias de repositorios

2. **NATS Subscriber**: Escucha mensajes del topic `inventario.cuadrilla`

3. **Worker Pool**: Pool de workers configurable usando Go routines
   - Procesamiento concurrente de mensajes
   - Configuración flexible del número de workers
   - Shutdown graceful con timeout

4. **Configuración**: Gestión de configuración mediante variables de entorno

## Estructura del Proyecto

```
service-worker-ps/
├── cmd/
│   └── service-worker-ps/
│       └── main.go              # Punto de entrada
├── internal/
│   ├── config/
│   │   └── config.go            # Gestión de configuración
│   ├── repository/
│   │   ├── repository.go        # Interface del repositorio
│   │   ├── postgres.go          # Implementación PostgreSQL
│   │   └── factory.go           # Factory pattern
│   ├── subscriber/
│   │   └── subscriber.go        # NATS subscriber
│   └── worker/
│       └── pool.go              # Worker pool
├── Dockerfile                   # Multi-stage build optimizado
├── go.mod
└── README.md
```

## Configuración

Variables de entorno disponibles:

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| NATS_URL | URL de conexión a NATS | nats://localhost:4222 |
| NATS_SUBJECT | Subject NATS para escuchar | inventario.cuadrilla |
| DATABASE_URL | String de conexión PostgreSQL | postgres://gridflow_user:gridflow_password@localhost:5432/gridflow?sslmode=disable |
| REPOSITORY_TYPE | Tipo de repositorio (postgresql, oracle, mongodb) | postgresql |
| WORKER_NUM_WORKERS | Número de workers en el pool | 10 |
| WORKER_BUFFER_SIZE | Tamaño del buffer de mensajes | 100 |
| WORKER_SHUTDOWN_TIMEOUT | Timeout de shutdown en segundos | 30 |

## Instalación y Ejecución

### Opción 1: Compilación Local

```bash
# Navegar al directorio del service worker
cd service-worker-ps

# Descargar dependencias
go mod download

# Compilar
go build -o service-worker-ps ./cmd/service-worker-ps

# Ejecutar
./service-worker-ps
```

### Opción 2: Docker (Recomendado)

```bash
# Desde el directorio raíz del repositorio
docker-compose up -d service-worker-ps

# Ver logs
docker-compose logs -f service-worker-ps

# Detener
docker-compose down
```

### Opción 3: Ejecutar Todos los Servicios

```bash
# Desde el directorio raíz del repositorio
docker-compose up -d

# Esto iniciará:
# - NATS (message broker)
# - PostgreSQL (base de datos)
# - gridflow-api (API REST)
# - service-worker-ps (service worker)
```

## Uso

El service worker se conecta automáticamente a NATS y comienza a escuchar mensajes en el topic configurado. Los mensajes recibidos se procesan de forma concurrente usando el worker pool y se persisten en PostgreSQL.

### Flujo de Datos

```
API REST → NATS (inventario.cuadrilla) → Service Worker → PostgreSQL
```

1. La API REST recibe solicitudes HTTP y publica eventos a NATS
2. El Service Worker escucha eventos de NATS
3. Los mensajes se distribuyen entre los workers del pool
4. Cada worker persiste el mensaje en PostgreSQL

## Extensibilidad

### Cambiar a Otra Base de Datos

Para cambiar de PostgreSQL a otra base de datos (por ejemplo, Oracle o MongoDB):

1. Implementar la interface `Repository` para la nueva base de datos:

```go
// oracle.go
type OracleRepository struct {
    // ...
}

func NewOracleRepository(connectionString string) (*OracleRepository, error) {
    // ...
}

func (r *OracleRepository) Save(ctx context.Context, data *InventarioData) error {
    // Implementación para Oracle
}

func (r *OracleRepository) Close() error {
    // ...
}

func (r *OracleRepository) HealthCheck(ctx context.Context) error {
    // ...
}
```

2. Actualizar el Factory en `factory.go`:

```go
case Oracle:
    return NewOracleRepository(connectionString)
```

3. Cambiar la variable de entorno:

```bash
export REPOSITORY_TYPE=oracle
export DATABASE_URL="oracle://user:pass@host:port/service"
```

## Características

- ✅ **Arquitectura extensible**: Patrón Repository para fácil cambio de base de datos
- ✅ **Compilación independiente**: No interfiere con el código existente
- ✅ **Worker pool configurable**: Procesamiento concurrente con Go routines
- ✅ **Shutdown graceful**: Manejo correcto de señales de terminación
- ✅ **Reconexión automática**: Manejo de desconexiones de NATS
- ✅ **Health checks**: Verificación de conexiones a servicios externos
- ✅ **Logging completo**: Información detallada de operaciones
- ✅ **Docker optimizado**: Imagen mínima (~15-20MB) con multi-stage build

## Métricas y Monitoreo

El service worker registra información importante:

- Conexión y reconexión a NATS
- Mensajes procesados exitosamente
- Errores de persistencia
- Estado de workers
- Shutdown graceful

## Desarrollo

### Agregar Tests

```bash
cd service-worker-ps
go test ./...
```

### Linting

```bash
go vet ./...
golangci-lint run
```

## Troubleshooting

### El service worker no se conecta a NATS

Verificar que NATS esté corriendo y accesible:

```bash
docker-compose logs nats
```

### Errores de conexión a PostgreSQL

Verificar la cadena de conexión y que PostgreSQL esté corriendo:

```bash
docker-compose logs postgres
```

### Workers no procesan mensajes

Verificar los logs del service worker:

```bash
docker-compose logs -f service-worker-ps
```

## Licencia

MIT License
