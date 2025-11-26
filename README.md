# GridFlow-Dynamics

Sistema de Seguimiento de Construcción de Red Eléctrica Basado en Eventos

## Descripción

GridFlow-Dynamics es una plataforma distribuida para monitorear en tiempo real la construcción y mantenimiento de infraestructura eléctrica utilizando arquitectura event-driven con RabbitMQ como backbone de mensajería. El sistema está diseñado para gestionar 200 cuadrillas simultáneas reportando ubicación, estado de tareas y alertas desde terreno.

## Arquitectura

La solución emplea microservicios desacoplados comunicados mediante eventos persistentes en RabbitMQ:

```
┌──────────────────────────────────────────────────────────────────┐
│                     GridFlow-Dynamics Platform                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │  Crew Tracking  │  │ Task Management │  │Alert Management │   │
│  │    Service      │  │     Service     │  │    Service      │   │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘   │
│           │                    │                    │            │
│  ┌────────┴────────────────────┴────────────────────┴────────┐   │
│  │                    REST API Layer                          │   │
│  │  POST /api/v1/tracking (Waze-like crew tracking)          │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                │                                 │
│                    ┌───────────▼───────────┐                     │
│                    │       RabbitMQ        │                     │
│                    │   (Event Backbone)    │                     │
│                    │                       │                     │
│                    │  ┌─────────────────┐  │                     │
│                    │  │  crew.events    │  │                     │
│                    │  │  crew.locations │  │                     │
│                    │  │  task.events    │  │                     │
│                    │  │  alert.events   │  │                     │
│                    │  └─────────────────┘  │                     │
│                    └───────────────────────┘                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Componentes

### Servicios

- **Crew Tracking Service**: Gestiona la ubicación y estado de las cuadrillas en tiempo real
- **Task Management Service**: Administra tareas de construcción, mantenimiento e inspección
- **Alert Management Service**: Procesa alertas de seguridad, equipos y logística desde terreno

### API REST - Tracking Tipo Waze

**Endpoint:** `POST /api/v1/tracking`

Recibe mensajes JSON cada 30 segundos desde aplicación móvil de campo.

#### Payload de Solicitud

```json
{
  "crewId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-15T10:30:00Z",
  "gpsCoordinates": {
    "lat": 40.7128,
    "lon": -74.0060
  },
  "taskId": "TASK-001",
  "status": "working",
  "progressPercentage": 75,
  "resourceConsumption": {
    "material_id": "MAT-001",
    "material_name": "Cable de cobre",
    "quantity": 150.5,
    "unit": "metros"
  },
  "safetyAlerts": [
    {
      "type": "warning",
      "description": "Condiciones de baja visibilidad",
      "severity": "medium"
    }
  ],
  "batteryLevel": 85,
  "region": "north"
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
  "success": true,
  "message": "Tracking update processed successfully",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
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

- `crewId`: UUID válido (formato 8-4-4-4-12)
- `timestamp`: ISO8601 válido
- `gpsCoordinates.lat`: -90 a 90
- `gpsCoordinates.lon`: -180 a 180
- `status`: en_route, working, paused, completed
- `progressPercentage`: 0-100
- `batteryLevel`: 0-100

### Eventos

| Exchange | Routing Keys | Descripción |
|----------|--------------|-------------|
| crew.events | crew.location.update, crew.status.update | Eventos de ubicación y estado de cuadrillas |
| crew.locations | crew.{region}.{crewId} | Eventos de tracking en tiempo real |
| task.events | task.status.update | Eventos de cambio de estado de tareas |
| alert.events | alert.created, alert.acknowledged, alert.resolved | Eventos del ciclo de vida de alertas |

### Modelos de Dominio

- **Crew**: Cuadrilla con ID, nombre, líder, ubicación GPS y estado
- **Task**: Tarea con tipo (construcción/mantenimiento/inspección/emergencia), prioridad y estado
- **Alert**: Alerta con categoría (seguridad/equipos/clima/logística), severidad y ubicación
- **TrackingPayload**: Datos de seguimiento desde app móvil

## Requisitos

- Go 1.21+
- RabbitMQ 3.12+

## Instalación

```bash
# Clonar el repositorio
git clone https://github.com/120m4n/GridFlow-Dynamics.git
cd GridFlow-Dynamics

# Descargar dependencias
go mod download

# Compilar
go build ./...
```

## Configuración

Variables de entorno:

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| RABBITMQ_URL | URL de conexión a RabbitMQ | amqp://guest:guest@localhost:5672/ |
| SERVER_PORT | Puerto del servidor | 8080 |
| HMAC_SECRET | Secreto para validación HMAC-SHA256 | default-secret-change-in-production |

## Ejecución

```bash
# Iniciar RabbitMQ (Docker)
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Ejecutar la plataforma
go run ./cmd/server
```

## Ejemplo de Uso - API Tracking

```bash
# Generar firma HMAC-SHA256
BODY='{"crewId":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2024-01-15T10:30:00Z","gpsCoordinates":{"lat":40.7128,"lon":-74.006},"taskId":"TASK-001","status":"working","progressPercentage":75,"batteryLevel":85}'
SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "your-secret-key" | awk '{print $2}')

# Enviar tracking update
curl -X POST http://localhost:8080/api/v1/tracking \
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
│   │   │   └── tracking.go      # Handler del endpoint de tracking
│   │   └── middleware/
│   │       ├── hmac.go          # Validación HMAC-SHA256
│   │       └── ratelimit.go     # Rate limiting por cuadrilla
│   ├── config/
│   │   └── config.go            # Gestión de configuración
│   ├── domain/
│   │   ├── crew.go              # Modelo de cuadrilla
│   │   ├── task.go              # Modelo de tarea
│   │   ├── alert.go             # Modelo de alerta
│   │   └── tracking.go          # Modelo de tracking
│   ├── messaging/
│   │   └── rabbitmq.go          # Infraestructura de mensajería
│   └── services/
│       ├── crew/
│       │   └── service.go       # Servicio de tracking de cuadrillas
│       ├── task/
│       │   └── service.go       # Servicio de gestión de tareas
│       └── alert/
│           └── service.go       # Servicio de gestión de alertas
├── go.mod
├── go.sum
└── README.md
```

## Capacidad

El sistema está diseñado para soportar:

- **200 cuadrillas simultáneas** reportando en tiempo real
- **Rate limiting**: 100 solicitudes/minuto por cuadrilla
- **Seguridad**: Validación HMAC-SHA256 en cada solicitud
- Eventos persistentes en RabbitMQ para garantizar entrega
- Arquitectura desacoplada para escalabilidad horizontal

## Licencia

MIT License
