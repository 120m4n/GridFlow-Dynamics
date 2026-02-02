# Documento de Diseño Arquitectónico para GridFlow-Dynamics

## Introducción
Este documento describe la arquitectura de persistencia de datos para el sistema GridFlow-Dynamics, que utiliza NATS como servicio de mensajería y PostgreSQL como base de datos para almacenar los datos enviados a través del endpoint `http://localhost:8080/api/v1/mensaje_inventario/cuadrilla`.

## Arquitectura General
La arquitectura del sistema se basa en un enfoque de microservicios, donde los datos de inventario de cuadrilla se envían a través de una API REST y se procesan mediante un servicio de mensajería (NATS). A continuación se detalla el flujo de datos:

1. **API REST**: La API recibe los mensajes de inventario de cuadrilla y los publica en NATS.
2. **NATS**: Actúa como un intermediario que permite la comunicación entre la API y el servicio de almacenamiento.
3. **Servicio de Almacenamiento**: Un servicio que escucha los mensajes de NATS y los almacena en una base de datos PostgreSQL.

## Conexión a NATS
Para establecer la conexión a NATS, se debe utilizar un cliente de NATS en el servicio de almacenamiento. A continuación se muestra un ejemplo de cómo se puede establecer la conexión:

```go
import (
    "github.com/nats-io/nats.go"
)

func connectToNATS() (*nats.Conn, error) {
    natsURL := "nats://localhost:4222"
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, err
    }
    return nc, nil
}
```

## Almacenamiento en PostgreSQL
Para almacenar los datos de inventario de cuadrilla, se debe crear una base de datos PostgreSQL y definir las tablas necesarias. A continuación se presenta un script para la creación de las tablas:

```sql
CREATE TABLE cuadrillas (
    id SERIAL PRIMARY KEY,
    grupo_trabajo VARCHAR(255) NOT NULL,
    nombre_empleado VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    latitud DECIMAL(9,6) NOT NULL,
    longitud DECIMAL(9,6) NOT NULL,
    codigo_odt VARCHAR(255) NOT NULL,
    estado VARCHAR(50) NOT NULL,
    porcentaje_progreso INT CHECK (porcentaje_progreso >= 0 AND porcentaje_progreso <= 100),
    nivel_bateria INT CHECK (nivel_bateria >= 0 AND nivel_bateria <= 100)
);
```

## Flujo de Datos
1. La API REST recibe un mensaje de inventario de cuadrilla.
2. La API publica el mensaje en NATS.
3. El servicio de almacenamiento escucha el mensaje en NATS.
4. El servicio de almacenamiento procesa el mensaje y lo inserta en la tabla `cuadrillas` de PostgreSQL.

## Conclusión
Este documento proporciona una visión general de la arquitectura de persistencia de datos para el sistema GridFlow-Dynamics, utilizando NATS y PostgreSQL. Se recomienda seguir las mejores prácticas de seguridad y rendimiento al implementar esta arquitectura.