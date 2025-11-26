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
│           └────────────────────┼────────────────────┘            │
│                                │                                 │
│                    ┌───────────▼───────────┐                     │
│                    │       RabbitMQ        │                     │
│                    │   (Event Backbone)    │                     │
│                    │                       │                     │
│                    │  ┌─────────────────┐  │                     │
│                    │  │  crew.events    │  │                     │
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

### Eventos

| Exchange | Routing Keys | Descripción |
|----------|--------------|-------------|
| crew.events | crew.location.update, crew.status.update | Eventos de ubicación y estado de cuadrillas |
| task.events | task.status.update | Eventos de cambio de estado de tareas |
| alert.events | alert.created, alert.acknowledged, alert.resolved | Eventos del ciclo de vida de alertas |

### Modelos de Dominio

- **Crew**: Cuadrilla con ID, nombre, líder, ubicación GPS y estado
- **Task**: Tarea con tipo (construcción/mantenimiento/inspección/emergencia), prioridad y estado
- **Alert**: Alerta con categoría (seguridad/equipos/clima/logística), severidad y ubicación

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

## Ejecución

```bash
# Iniciar RabbitMQ (Docker)
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Ejecutar la plataforma
go run ./cmd/server
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
│       └── main.go          # Punto de entrada principal
├── internal/
│   ├── config/
│   │   └── config.go        # Gestión de configuración
│   ├── domain/
│   │   ├── crew.go          # Modelo de cuadrilla
│   │   ├── task.go          # Modelo de tarea
│   │   └── alert.go         # Modelo de alerta
│   ├── messaging/
│   │   └── rabbitmq.go      # Infraestructura de mensajería
│   └── services/
│       ├── crew/
│       │   └── service.go   # Servicio de tracking de cuadrillas
│       ├── task/
│       │   └── service.go   # Servicio de gestión de tareas
│       └── alert/
│           └── service.go   # Servicio de gestión de alertas
├── go.mod
├── go.sum
└── README.md
```

## Capacidad

El sistema está diseñado para soportar:

- **200 cuadrillas simultáneas** reportando en tiempo real
- Eventos persistentes en RabbitMQ para garantizar entrega
- Arquitectura desacoplada para escalabilidad horizontal

## Licencia

MIT License
