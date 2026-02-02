# Mensaje de Inventario de Cuadrilla

Este mensaje contiene información relevante sobre el estado y progreso de una cuadrilla en una tarea específica.


```json
{
  "grupoTrabajo": "G0/CUADRILLA_123",
  "nombreEmpleado": "Juan Perez",
  "timestamp": "2024-01-15T10:30:00Z",
  "coordenadas": {
    "latitud": 40.7128,
    "longitud": -74.0060
  }
  "codigoODT": "codigoodt_consecutivo",
  "estado": "trabajando",
  "procentajeProgreso": 75,
  "nivelBateria": 85
}
```

## Endpoint
`POST /api/v1/mensaje_inventario/cuadrilla`

## Calculo de Firma HMAC-SHA256
La firma HMAC-SHA256 se calcula utilizando una clave secreta compartida y el cuerpo del mensaje en formato JSON. La firma se incluye en el header `X-Signature-256`.

ejemplo en pseudocódigo:
```
clave_secreta = "tu_clave_secreta"
body_json = obtener_cuerpo_json()
firma = HMAC_SHA256(clave_secreta, body_json)
```



## Headers Requeridos
|Header |Descripción|
|:--------:|:---------:|
|X-Signature-256|Firma HMAC-SHA256 del body|
|Content-Type|application/json|

## Descripción de Campos
- `grupoTrabajo`: Grupo administrativo mas codigo grupo trabajo .
- `nombreEmpleado`: Nombre completo del empleado que envía el mensaje.
- `timestamp`: Marca de tiempo del mensaje en formato ISO 8601.
- `coordenadas`: Objeto que contiene la latitud y longitud actuales de la cuadrilla.
- `codigoODT`: Código único de la orden de trabajo + consecutivo (odt hija).
- `estado`: Estado actual de la cuadrilla (e.g., "trabajando", "en pausa", "finalizado").
- `procentajeProgreso`: Porcentaje de progreso de la tarea asignada.
- `nivelBateria`: Nivel de batería del dispositivo utilizado por la cuadrilla, expresado en porcentaje. 

## Códigos de Error
|Código |Descripción|
|--------|------|
|400|Payload inválido o campos faltantes|
|401|Firma HMAC-SHA256 inválida o faltante
|405|Método no permitido (solo POST)
|429|Rate limit excedido (100 req/min)
|500|Error interno del servidor

## Validaciones
grupoTrabajo: cadena no vacía
nombreEmpleado: cadena no vacía
codigoODT: cadena no vacía
timestamp: ISO8601 válido
coordenadas.latitud: -90 a 90
coordenadas.longitud: -180 a 180
estado: en_ruta, trabajando, en_pausa, finalizado (nota: si no puede ser calculado, usar "trabajando")
procentajeProgreso: 0-100 (nota: si no puede ser calculado, usar 100)
nivelBateria: 0-100

## Ejemplo de Respuesta Exitosa
```json
{
  "status": "success",
  "message": "Mensaje de inventario de cuadrilla recibido correctamente."
}
```

## Revision
Autor: Equipo de Desarrollo API
Versión: 1.0.0