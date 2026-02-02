# GridFlow-Dynamics

Sistema de Seguimiento de Construcción de Red Eléctrica Basado en Eventos

## Descripción

GridFlow-Dynamics es una plataforma distribuida para monitorear en tiempo real la construcción y mantenimiento de infraestructura eléctrica utilizando arquitectura event-driven con NATS como backbone de mensajería. El sistema está diseñado para gestionar 200 cuadrillas simultáneas reportando inventario y progreso desde terreno.

## Arquitectura

La solución emplea una API REST que publica eventos a NATS:

```
┌──────────────────────────────────────────────────────────────────┐
│                     GridFlow-Dynamics Platform                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │                    REST API Layer                          │   │
│  │  POST /api/v1/mensaje_inventario/cuadrilla                 │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                │                                 │
│                    ┌───────────▼───────────┐                     │
│                    │         NATS          │                     │
│                    │   (Event Backbone)    │                     │
│                    │                       │                     │
│                    │  ┌─────────────────┐  │                     │
│                    │  │ inventario.*    │  │                     │
│                    │  └─────────────────┘  │                     │
│                    └───────────────────────┘                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Componentes

### API REST - Mensaje de Inventario de Cuadrilla

**Endpoint:** `POST /api/v1/mensaje_inventario/cuadrilla`

Recibe mensajes JSON desde aplicación móvil de campo con inventario y progreso de la cuadrilla.

#### Payload de Solicitud

```json
{
  "grupoTrabajo": "G0/CUADRILLA_123",
  "nombreEmpleado": "Juan Perez",
  "timestamp": "2024-01-15T10:30:00Z",
  "coordenadas": {
    "latitud": 40.7128,
    "longitud": -74.0060
  },
  "codigoODT": "codigoodt_consecutivo",
  "estado": "trabajando",
  "procentajeProgreso": 75,
  "nivelBateria": 85
}
```

#### Headers Requeridos

| Header | Descripción |
|--------|-------------|
| X-Signature-256 | Firma HMAC-SHA256 del body |
| Content-Type | application/json |

#### Respuesta Exitosa (200)

```json
{
  "status": "success",
  "message": "Mensaje de inventario de cuadrilla recibido correctamente."
}
```

#### Códigos de Error

| Código | Descripción |
|--------|-------------|
| 400 | Payload inválido o campos faltantes |
| 401 | Firma HMAC-SHA256 inválida o faltante |
| 405 | Método no permitido (solo POST) |
| 429 | Rate limit excedido (100 req/min) |
| 500 | Error interno del servidor |

#### Validaciones

- `grupoTrabajo`: cadena no vacía
- `nombreEmpleado`: cadena no vacía
- `codigoODT`: cadena no vacía
- `timestamp`: ISO8601 válido
- `coordenadas.latitud`: -90 a 90
- `coordenadas.longitud`: -180 a 180
- `estado`: en_ruta, trabajando, en_pausa, finalizado
- `procentajeProgreso`: 0-100
- `nivelBateria`: 0-100

### Eventos

| Subject | Descripción |
|---------|-------------|
| inventario.cuadrilla | Evento de inventario de cuadrilla publicado por la API |

### Modelo de Dominio

- **MensajeInventarioCuadrilla**: Datos de inventario y progreso desde la app móvil

## Requisitos

- Go 1.21+
- NATS 2.x+
- Docker y Docker Compose (para deployment)
- PostgreSQL 15+ (para persistencia de datos)

## Instalación

### Opción 1: Desarrollo Local

```bash
# Clonar el repositorio
git clone https://github.com/120m4n/GridFlow-Dynamics.git
cd GridFlow-Dynamics

# Descargar dependencias
go mod download

# Compilar
go build ./...
```

### Opción 2: Docker (Recomendado)

```bash
# Clonar el repositorio
git clone https://github.com/120m4n/GridFlow-Dynamics.git
cd GridFlow-Dynamics

# Construir y ejecutar con Docker Compose
docker-compose up -d

# Ver logs
docker-compose logs -f gridflow-api

# Detener servicios
docker-compose down
```

## Configuración

Variables de entorno:

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| NATS_URL | URL de conexión a NATS | nats://localhost:4222 |
| SERVER_PORT | Puerto del servidor | 8080 |
| HMAC_SECRET | Secreto para validación HMAC-SHA256 | default-secret-change-in-production |

## Ejecución

### Desarrollo Local

```bash
# Iniciar NATS (Docker)
docker run -d --name nats -p 4222:4222 nats:2

# Ejecutar la plataforma
go run ./cmd/server
```

### Producción con Docker

```bash
# Configurar secreto HMAC (opcional)
export HMAC_SECRET="your-production-secret-key"

# Iniciar todos los servicios (API, NATS, PostgreSQL)
docker-compose up -d

# Verificar estado de los servicios
docker-compose ps

# Ver logs de la API
docker-compose logs -f gridflow-api

# Verificar health de la API
curl http://localhost:8080/health
```

### Build de Imagen Docker

La imagen Docker está optimizada para tamaño mínimo usando multi-stage build:

```bash
# Construir imagen manualmente
docker build -t gridflow-dynamics:latest .

# Ver tamaño de la imagen (aproximadamente 15-20MB)
docker images gridflow-dynamics

# Ejecutar solo la API
docker run -d \
  --name gridflow-api \
  -p 8080:8080 \
  -e NATS_URL=nats://host.docker.internal:4222 \
  -e HMAC_SECRET=your-secret \
  gridflow-dynamics:latest
```

**Optimizaciones de tamaño de imagen:**
- Multi-stage build con `golang:1.21-alpine` y `scratch`
- Binario compilado estáticamente sin CGO
- Símbolos de debug eliminados (`-ldflags="-w -s"`)
- Tamaño final: ~15-20MB

## Ejemplo de Uso - API Mensaje Inventario

```bash
# Generar firma HMAC-SHA256
BODY='{"grupoTrabajo":"G0/CUADRILLA_123","nombreEmpleado":"Juan Perez","timestamp":"2024-01-15T10:30:00Z","coordenadas":{"latitud":40.7128,"longitud":-74.006},"codigoODT":"codigoodt_consecutivo","estado":"trabajando","procentajeProgreso":75,"nivelBateria":85}'
SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "your-secret-key" | awk '{print $2}')

# Enviar mensaje de inventario
curl -X POST http://localhost:8080/api/v1/mensaje_inventario/cuadrilla \
  -H "Content-Type: application/json" \
  -H "X-Signature-256: $SIGNATURE" \
  -d "$BODY"
```

## Pruebas

```bash
# Ejecutar todas las pruebas
go test ./...

# Ejecutar pruebas con cobertura
go test ./... -cover

# Ejecutar pruebas verbose
go test ./... -v
```

## Estructura del Proyecto

```
GridFlow-Dynamics/
├── cmd/
│   └── server/
│       └── main.go              # Punto de entrada principal
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   └── tracking.go      # Handler del endpoint de inventario
│   │   └── middleware/
│   │       ├── hmac.go          # Validación HMAC-SHA256
│   │       └── ratelimit.go     # Rate limiting por cuadrilla
│   ├── config/
│   │   └── config.go            # Gestión de configuración
│   ├── domain/
│   │   └── tracking.go          # Modelo de inventario de cuadrilla
│   └── messaging/
│       └── nats.go              # Infraestructura de mensajería
├── scripts/
│   └── init.sql                 # Script de inicialización PostgreSQL
├── Dockerfile                   # Multi-stage build optimizado
├── docker-compose.yml           # Orquestación de servicios
├── .dockerignore                # Exclusiones para build Docker
├── go.mod
├── go.sum
└── README.md
```

## Docker Compose

El archivo `docker-compose.yml` incluye:

- **gridflow-api**: API REST (puerto 8080)
- **nats**: Message broker (puertos 4222, 8222, 6222)
- **postgres**: Base de datos (puerto 5432)

Todos los servicios incluyen healthchecks y reinicio automático.

## Capacidad

El sistema está diseñado para soportar:

- **200 cuadrillas simultáneas** reportando en tiempo real
- **Rate limiting**: 100 solicitudes/minuto por cuadrilla
- **Seguridad**: Validación HMAC-SHA256 en cada solicitud
- Eventos publicados en NATS para integración con consumidores externos
- Arquitectura desacoplada para escalabilidad horizontal

## Licencia

MIT License
